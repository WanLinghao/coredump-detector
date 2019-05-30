/*
Copyright 2017 The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/WanLinghao/fujitsu-coredump/pkg/apis/coredump/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CoredumpEndpointLister helps list CoredumpEndpoints.
type CoredumpEndpointLister interface {
	// List lists all CoredumpEndpoints in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.CoredumpEndpoint, err error)
	// CoredumpEndpoints returns an object that can list and get CoredumpEndpoints.
	CoredumpEndpoints(namespace string) CoredumpEndpointNamespaceLister
	CoredumpEndpointListerExpansion
}

// coredumpEndpointLister implements the CoredumpEndpointLister interface.
type coredumpEndpointLister struct {
	indexer cache.Indexer
}

// NewCoredumpEndpointLister returns a new CoredumpEndpointLister.
func NewCoredumpEndpointLister(indexer cache.Indexer) CoredumpEndpointLister {
	return &coredumpEndpointLister{indexer: indexer}
}

// List lists all CoredumpEndpoints in the indexer.
func (s *coredumpEndpointLister) List(selector labels.Selector) (ret []*v1alpha1.CoredumpEndpoint, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CoredumpEndpoint))
	})
	return ret, err
}

// CoredumpEndpoints returns an object that can list and get CoredumpEndpoints.
func (s *coredumpEndpointLister) CoredumpEndpoints(namespace string) CoredumpEndpointNamespaceLister {
	return coredumpEndpointNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CoredumpEndpointNamespaceLister helps list and get CoredumpEndpoints.
type CoredumpEndpointNamespaceLister interface {
	// List lists all CoredumpEndpoints in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.CoredumpEndpoint, err error)
	// Get retrieves the CoredumpEndpoint from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.CoredumpEndpoint, error)
	CoredumpEndpointNamespaceListerExpansion
}

// coredumpEndpointNamespaceLister implements the CoredumpEndpointNamespaceLister
// interface.
type coredumpEndpointNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CoredumpEndpoints in the indexer for a given namespace.
func (s coredumpEndpointNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.CoredumpEndpoint, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CoredumpEndpoint))
	})
	return ret, err
}

// Get retrieves the CoredumpEndpoint from the indexer for a given namespace and name.
func (s coredumpEndpointNamespaceLister) Get(name string) (*v1alpha1.CoredumpEndpoint, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("coredumpendpoint"), name)
	}
	return obj.(*v1alpha1.CoredumpEndpoint), nil
}
