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

package backend

import (
	"fmt"
	"sync"


	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/options"
	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/types"
	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/volume"
)

var backendStorage types.Storage
var mu sync.Mutex

func GetBackendStorage() types.Storage {
	mu.Lock()
	defer mu.Unlock()
	var err error

	if backendStorage == nil  {
		// initialize backend storage 
		if options.BackendOpts.BackendStorageKind == "local" {
			backendStorage, err = volume.NewLocalStorage(options.BackendOpts.LocalPath)
			if err != nil {
				panic(err)
			}
		} else if options.BackendOpts.BackendStorageKind == "aws" {
			backendStorage, err = volume.NewAwsStorage(options.BackendOpts.AwsS3Host, options.BackendOpts.AwsS3AccessKey,
				options.BackendOpts.AwsS3SecretKey, options.BackendOpts.AwsS3Region, options.BackendOpts.AwsS3Bucket, true)
			if err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("unsupportedd volume type:%s, only support 'aws' or 'local'", options.BackendOpts.BackendStorageKind))
		}
	}
	return backendStorage
}

// func init() {
// 	var (
// 		s   types.Storage
// 		err error
// 	)
// 	if options.BackendOpts.BackendStorageKind == "local" {
// 		s, err = volume.NewLocalStorage(options.BackendOpts.LocalPath)
// 		if err != nil {
// 			panic(err)
// 		}
// 	} else if options.BackendOpts.BackendStorageKind == "aws" {
// 		s, err = volume.NewAwsStorage(options.BackendOpts.AwsS3Host, options.BackendOpts.AwsS3AccessKey,
// 			options.BackendOpts.AwsS3SecretKey, options.BackendOpts.AwsS3Region, options.BackendOpts.AwsS3Bucket, true)
// 		if err != nil {
// 			panic(err)
// 		}
// 	} else {
// 		panic(fmt.Errorf("unsupported 	:%s, only support 'aws' or 'local'", options.BackendOpts.BackendStorageKind))
// 	}
// 	BackendStorage = s
// }
