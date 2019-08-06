/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	//	"k8s.io/api/core/v1"
	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/types"
	//	"github.com/WanLinghao/fujitsu-coredump/pkg/backend"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog"
	"time"
)

type localStorage struct {
	metadataFile string
	rootPath     string
}

func NewLocalStorage(rootPath string) (types.Storage, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("the root path for local storage can't be empty")
	}
	return &localStorage{
		metadataFile: "metadata.json",
		rootPath:     rootPath}, nil
}

func (l *localStorage) GetCoreFiles(ns, podUID, container string) (string, error) {
	coreDirPath, err := l.getPath(ns, podUID, container)
	if err != nil {
		return "", err
	}
	klog.V(1).Infof("we are tar %s\n", coreDirPath)

	stat, err := os.Stat(coreDirPath)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("the %s/%s has no core files in namespace %s", podUID, container, ns)
		} else {
			return "", fmt.Errorf("got unexpected error when access %s: %v", coreDirPath, err)
		}
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("got unexpected error: %s is a not a directory", coreDirPath)
	}

	return makeTar(coreDirPath, "/tmp")
}

func (l *localStorage) CleanNamespace(namespace string) error {
	path, err := l.getPath(namespace, "", "")
	if err != nil {
		return fmt.Errorf("clean namespace %s failed: %v", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("clean namespace %s failed: %v", err)
	}
	return nil
}

func (l *localStorage) LogPodDeletion(namespace string, podUID string, gcTimeStamp time.Time) error {
	path, err := l.getPath(namespace, podUID, "")
	if err != nil {
		return fmt.Errorf("clean pod %s/%s failed: %v", namespace, podUID, err)
	}

	err = marshal(&types.Metadata{
		GCTimeStamp: &gcTimeStamp}, path)
	if err != nil {
		return fmt.Errorf("log deletion timestamp for pod %s/%s", namespace, podUID, err)
	}

	return nil
}

func (l *localStorage) GC() error {
	dirs, err := ioutil.ReadDir(l.rootPath)

	if err != nil {
		return err
	}

	allErrors := []error{}
	for _, dir := range dirs {
		if dir.IsDir() {
			namespaceDir, err := l.getPath(dir.Name(), "", "")
			if err != nil {
				allErrors = append(allErrors, err)
				continue
			}

			dirs2, err := ioutil.ReadDir(namespaceDir)
			if err != nil {
				allErrors = append(allErrors, err)
				continue
			}

			for _, dir2 := range dirs2 {
				podDir, err := l.getPath(dir2.Name(), "", "")
				if err != nil {
					allErrors = append(allErrors, err)
					continue
				}
				metaPath := podDir + "/" + l.metadataFile
				_, err = os.Stat(metaPath)
				if err != nil {
					if os.IsExist(err) {
						allErrors = append(allErrors, err)
					}
					continue
				}
				md, err := unmarshal(metaPath)
				if err != nil {
					allErrors = append(allErrors, err)
					continue
				}
				if md.GCTimeStamp != nil {
					time.Now().After(*md.GCTimeStamp)
					err = os.RemoveAll(podDir)
					if err != nil {
						allErrors = append(allErrors, err)
						continue
					}
				}
			}
		}
	}
	return utilerrors.NewAggregate(allErrors)
}

// // This function handles clean job of core files by namespace and podUID
// // If podUID == "", then it means clean all the core files beneath that namespace
// func (l *localStorage) CleanCoreFiles(ns, podUID string) error {
// 	path := l.getPath(ns, podUID, "")
// 	if path == "" {
// 		// indicates the path is illegal
// 		return fmt.Errorf("try clean a illegal path, namespace is %s, podUID is %s", ns, podUID)
// 	}

// 	return os.Remove(path)
// }

// getPath returns a absolute path, if the path is not exist, it returns ""
func (l *localStorage) getPath(ns, podUID, container string) (string, error) {
	ret := l.rootPath + "/" + ns + "/" + podUID + "/" + container
	ret = filepath.Clean(ret)
	if ret == l.rootPath || !strings.HasPrefix(ret, l.rootPath) {
		// We should not delete root path or any non-sub directory of root path
		return "", fmt.Errorf("failed get path for %s/%s/%s: illegal path '%s'", ns, podUID, container, ret)
	}

	_, err := os.Stat(ret)
	if err != nil {
		// It includes non-exist error
		return "", fmt.Errorf("failed get path for %s/%s/%s: %v", ns, podUID, container, err)
	}
	return ret, nil
}

func unmarshal(filepath string) (*types.Metadata, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// m stores the metadata information
	m := &types.Metadata{}
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func marshal(m *types.Metadata, filepath string) error {
	data, err := json.MarshalIndent(*m, "", " ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath, data, 0644)
}
