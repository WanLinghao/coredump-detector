package sqllog

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/rancher/kine/pkg/broadcaster"
	"github.com/rancher/kine/pkg/server"
	"github.com/sirupsen/logrus"
)

type SQLLog struct {
	d           Dialect
	broadcaster broadcaster.Broadcaster
	ctx         context.Context
	notify      chan int64
}

func New(d Dialect) *SQLLog {
	l := &SQLLog{
		d:      d,
		notify: make(chan int64, 1024),
	}
	return l
}

type Dialect interface {
	ListCurrent(ctx context.Context, prefix string, limit int64, includeDeleted bool) (*sql.Rows, error)
	List(ctx context.Context, prefix, startKey string, limit, revision int64, includeDeleted bool) (*sql.Rows, error)
	Count(ctx context.Context, prefix string) (int64, int64, error)
	CurrentRevision(ctx context.Context) (int64, error)
	After(ctx context.Context, prefix string, rev int64) (*sql.Rows, error)
	Insert(ctx context.Context, key string, create, delete bool, createRevision, previousRevision int64, ttl int64, value, prevValue []byte) (int64, error)
	GetRevision(ctx context.Context, revision int64) (*sql.Rows, error)
	DeleteRevision(ctx context.Context, revision int64) error
	GetCompactRevision(ctx context.Context) (int64, error)
	SetCompactRevision(ctx context.Context, revision int64) error
}

func (s *SQLLog) Start(ctx context.Context) error {
	s.ctx = ctx
	go s.compact()
	return nil
}

func (s *SQLLog) compact() {
	t := time.NewTicker(2 * time.Second)

outer:
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
		}

		end, err := s.d.CurrentRevision(s.ctx)
		if err != nil {
			logrus.Errorf("failed to get current revision: %v", err)
			continue
		}

		cursor, err := s.d.GetCompactRevision(s.ctx)
		if err != nil {
			logrus.Errorf("failed to get compact revision: %v", err)
			continue
		}

		if end-cursor < 100 {
			// Only run if we have at least 100 rows to process
			continue
		}

		savedCursor := cursor
		// Purposefully start at the current and redo the current as
		// it could have failed before actually compacting
		for ; cursor <= end; cursor++ {
			rows, err := s.d.GetRevision(s.ctx, cursor)
			if err != nil {
				logrus.Errorf("failed to get revision %d: %v", cursor, err)
				continue outer
			}

			_, _, events, err := RowsToEvents(rows)
			if err != nil {
				logrus.Errorf("failed to convert to events: %v", err)
				continue outer
			}

			if len(events) == 0 {
				continue
			}

			event := events[0]

			if event.KV.Key == "compact_rev_key" {
				// don't compact the compact key
				continue
			}

			setRev := false
			if event.PrevKV != nil && event.PrevKV.ModRevision != 0 {
				if savedCursor != cursor {
					if err := s.d.SetCompactRevision(s.ctx, cursor); err != nil {
						logrus.Errorf("failed to record compact revision: %v", err)
						continue outer
					}
					savedCursor = cursor
					setRev = true
				}

				if err := s.d.DeleteRevision(s.ctx, event.PrevKV.ModRevision); err != nil {
					logrus.Errorf("failed to delete revision %d: %v", event.PrevKV.ModRevision, err)
					continue outer
				}
			}

			if event.Delete {
				if !setRev && savedCursor != cursor {
					if err := s.d.SetCompactRevision(s.ctx, cursor); err != nil {
						logrus.Errorf("failed to record compact revision: %v", err)
						continue outer
					}
					savedCursor = cursor
				}

				if err := s.d.DeleteRevision(s.ctx, cursor); err != nil {
					logrus.Errorf("failed to delete current revision %d: %v", cursor, err)
					continue outer
				}
			}
		}

		if savedCursor != cursor {
			if err := s.d.SetCompactRevision(s.ctx, cursor); err != nil {
				logrus.Errorf("failed to record compact revision: %v", err)
				continue outer
			}
		}
	}
}

func (s *SQLLog) CurrentRevision(ctx context.Context) (int64, error) {
	return s.d.CurrentRevision(ctx)
}

func (s *SQLLog) After(ctx context.Context, prefix string, revision int64) (int64, []*server.Event, error) {
	if strings.HasSuffix(prefix, "/") {
		prefix += "%"
	}

	rows, err := s.d.After(ctx, prefix, revision)
	if err != nil {
		return 0, nil, err
	}

	rev, _, result, err := RowsToEvents(rows)
	return rev, result, err
}

func (s *SQLLog) List(ctx context.Context, prefix, startKey string, limit, revision int64, includeDeleted bool) (int64, []*server.Event, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if strings.HasSuffix(prefix, "/") {
		prefix += "%"
	}

	if revision == 0 {
		rows, err = s.d.ListCurrent(ctx, prefix, limit, includeDeleted)
	} else {
		rows, err = s.d.List(ctx, prefix, startKey, limit, revision, includeDeleted)
	}
	if err != nil {
		return 0, nil, err
	}

	rev, compact, result, err := RowsToEvents(rows)
	if err != nil {
		return 0, nil, err
	}

	if revision > 0 && len(result) == 0 {
		// a zero length result won't have the compact revision so get it manually
		compact, err = s.d.GetCompactRevision(ctx)
		if err != nil {
			return 0, nil, err
		}
	}

	if revision > 0 && revision < compact {
		return rev, result, server.ErrCompacted
	}

	select {
	case s.notify <- rev:
	default:
	}

	return rev, result, err
}

func RowsToEvents(rows *sql.Rows) (int64, int64, []*server.Event, error) {
	var (
		result  []*server.Event
		rev     int64
		compact int64
	)
	defer rows.Close()

	for rows.Next() {
		event := &server.Event{}
		if err := scan(rows, &rev, &compact, event); err != nil {
			return 0, 0, nil, err
		}
		result = append(result, event)
	}

	return rev, compact, result, nil
}

func (s *SQLLog) Watch(ctx context.Context, prefix string) <-chan []*server.Event {
	res := make(chan []*server.Event)
	values, err := s.broadcaster.Subscribe(ctx, s.startWatch)
	if err != nil {
		return nil
	}

	checkPrefix := strings.HasSuffix(prefix, "/")

	go func() {
		defer close(res)
		for i := range values {
			events, ok := filter(i, checkPrefix, prefix)
			if ok {
				res <- events
			}
		}
	}()

	return res
}

func filter(events interface{}, checkPrefix bool, prefix string) ([]*server.Event, bool) {
	eventList := events.([]*server.Event)
	filteredEventList := make([]*server.Event, 0, len(eventList))

	for _, event := range eventList {
		if (checkPrefix && strings.HasPrefix(event.KV.Key, prefix)) || event.KV.Key == prefix {
			filteredEventList = append(filteredEventList, event)
		}
	}

	return filteredEventList, len(filteredEventList) > 0
}

func (s *SQLLog) startWatch() (chan interface{}, error) {
	c := make(chan interface{})
	go s.poll(c)
	return c, nil
}

func (s *SQLLog) poll(result chan interface{}) {
	var (
		last int64
	)

	wait := time.NewTicker(120 * time.Second)
	defer wait.Stop()
	defer close(result)

	for {
		select {
		case <-s.ctx.Done():
			return
		case check := <-s.notify:
			if check <= last {
				continue
			}
		case <-wait.C:
		}

		rows, err := s.d.After(s.ctx, "%", last)
		if err != nil {
			logrus.Errorf("fail to list latest changes: %v", err)
			continue
		}

		rev, _, events, err := RowsToEvents(rows)
		if err != nil {
			logrus.Errorf("fail to convert rows changes: %v", err)
			continue
		}

		if len(events) == 0 {
			continue
		}

		for _, event := range events {
			logrus.Debugf("TRIGGERED %s, revision=%d, delete=%v", event.KV.Key, event.KV.ModRevision, event.Delete)
		}

		result <- events
		last = rev
	}
}

func (s *SQLLog) Count(ctx context.Context, prefix string) (int64, int64, error) {
	if strings.HasSuffix(prefix, "/") {
		prefix += "%"
	}
	return s.d.Count(ctx, prefix)
}

func (s *SQLLog) Append(ctx context.Context, event *server.Event) (int64, error) {
	e := *event
	if e.KV == nil {
		e.KV = &server.KeyValue{}
	}
	if e.PrevKV == nil {
		e.PrevKV = &server.KeyValue{}
	}

	rev, err := s.d.Insert(ctx, e.KV.Key,
		e.Create,
		e.Delete,
		e.KV.CreateRevision,
		e.PrevKV.ModRevision,
		e.KV.Lease,
		e.KV.Value,
		e.PrevKV.Value,
	)
	if err != nil {
		return 0, err
	}
	select {
	case s.notify <- rev:
	default:
	}
	return rev, nil
}

func scan(rows *sql.Rows, rev *int64, compact *int64, event *server.Event) error {
	event.KV = &server.KeyValue{}
	event.PrevKV = &server.KeyValue{}

	c := &sql.NullInt64{}

	err := rows.Scan(
		rev,
		c,
		&event.KV.ModRevision,
		&event.KV.Key,
		&event.Create,
		&event.Delete,
		&event.KV.CreateRevision,
		&event.PrevKV.ModRevision,
		&event.KV.Lease,
		&event.KV.Value,
		&event.PrevKV.Value,
	)
	if err != nil {
		return err
	}

	if event.Create {
		event.KV.CreateRevision = event.KV.ModRevision
		event.PrevKV = nil
	}

	*compact = c.Int64
	return nil
}
