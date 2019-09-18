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

package main

import (
	//	"time"
	"github.com/spf13/cobra"

	backendopts "github.com/WanLinghao/coredump-detector/pkg/backend/options"
	"github.com/WanLinghao/coredump-detector/pkg/gc"
	gcopts "github.com/WanLinghao/coredump-detector/pkg/gc/options"

	genericapiserver "k8s.io/apiserver/pkg/server"
)

func main() {
	signalCh := genericapiserver.SetupSignalHandler()

	cmd := &cobra.Command{
		Short: "Launch an API server",
		Long:  "Launch an API server",
		RunE: func(c *cobra.Command, args []string) error {
			bgc, err := gc.NewBackendGC()
			if err != nil {
				return err
			}
			bgc.GC(signalCh)
			<-signalCh
			return nil
		},
	}

	err := backendopts.BackendPathSetFunc(cmd)
	if err != nil {
		panic(err)
	}

	err = gcopts.GCSetFunc(cmd)
	if err != nil {
		panic(err)
	}

	if err = cmd.Execute(); err != nil {
		panic(err)
	}
}
