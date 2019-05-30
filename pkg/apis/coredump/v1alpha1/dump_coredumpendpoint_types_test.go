
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



package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/WanLinghao/fujitsu-coredump/pkg/apis/coredump/v1alpha1"
	. "github.com/WanLinghao/fujitsu-coredump/pkg/client/clientset_generated/clientset/typed/coredump/v1alpha1"
)

var _ = Describe("CoredumpEndpoint", func() {
	var instance CoredumpEndpoint
	var expected CoredumpEndpoint
	var client CoredumpEndpointInterface

	BeforeEach(func() {
		instance = CoredumpEndpoint{}
		instance.Name = "instance-1"

		expected = instance
	})

	AfterEach(func() {
		client.Delete(instance.Name, &metav1.DeleteOptions{})
	})

	Describe("when sending a dump request", func() {
		It("should return success", func() {
			client = cs.CoredumpV1alpha1().Coredumpendpoints("coredumpendpoint-test-dump")
			_, err := client.Create(&instance)
			Expect(err).ShouldNot(HaveOccurred())

			dump := &CoredumpEndpointDump{}
			dump.Name = instance.Name
			restClient := cs.CoredumpV1alpha1().RESTClient()
			err = restClient.Post().Namespace("coredumpendpoint-test-dump").
				Name(instance.Name).
				Resource("coredumpendpoints").
				SubResource("dump").
				Body(dump).Do().Error()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
