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

package main

import (
	// Make sure dep tools picks up these dependencies
	_ "github.com/go-openapi/loads"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth" // Enable cloud provider auth
	"sigs.k8s.io/apiserver-builder-alpha/pkg/cmd/server"

	"github.com/WanLinghao/coredump-detector/pkg/apis"
	backend "github.com/WanLinghao/coredump-detector/pkg/backend/options"
	"github.com/WanLinghao/coredump-detector/pkg/openapi"
	"github.com/spf13/cobra"
)

func main() {
	version := "v0"

	flagFuncs := []func(*cobra.Command) error{
		backend.BackendPathSetFunc,
	}

	err := server.StartApiServerWithOptions(&server.StartOptions{
		EtcdPath:        "/registry/fujitsu.com",
		Apis:            apis.GetAllApiBuilders(),
		Openapidefs:     openapi.GetOpenAPIDefinitions,
		Title:           "Api",
		Version:         version,
		FlagConfigFuncs: flagFuncs,
		// TweakConfigFuncs []func(apiServer *apiserver.Config) error
		// FlagConfigFuncs []func(*cobra.Command) error
	})
	if err != nil {
		panic(err)
	}
}
