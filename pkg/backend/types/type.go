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

package types

import (
	"time"
)

type BackendOptions struct {
	BackendStorageKind string

	// Settings for aws storage
	AwsS3Host      string
	AwsS3AccessKey string
	AwsS3SecretKey string
	AwsS3Region    string
	AwsS3Bucket    string

	// Settings for local storage
	LocalPath string
}

type Storage interface {
	GetCoreFiles(namespace string, pod string, container string) (string, error)
	CleanNamespace(namespace string) error
	LogPodDeletion(namespace string, podUID string, deletionTimestamp time.Time) error
	GC() error
	//ListNamespacesPods() map[string][]string
}

type Metadata struct {
	GCTimeStamp   *time.Time `json:"gcTimeStamp"`
	DownloadCount int        `json:"donwloadCount"`
}
