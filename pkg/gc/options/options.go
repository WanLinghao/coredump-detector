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

package options

import (
	"github.com/spf13/cobra"
	"time"
)

type GCOptions struct {
	GCPeriod    time.Duration
	GCThreshold time.Duration
	//	CoreFilesMaxSizePerContainer string
}

var (
	GCOpts    *GCOptions
	GCSetFunc func(*cobra.Command) error
)

func init() {
	GCOpts = &GCOptions{}
	GCSetFunc = func(cmd *cobra.Command) error {
		flags := cmd.Flags()
		flags.DurationVar(&GCOpts.GCPeriod, "gc-period", 1*time.Minute, "gc period")
		flags.DurationVar(&GCOpts.GCThreshold, "gc-threshold", 24*time.Hour, "gc threshold, any resource terminated after this duration would be cleaned")
		//	flags.StringVar(&GCOpts.CoreFilesMaxSizePerContainer, "max-size-per-container", "1GB", "max size per container")
		return nil
	}
}
