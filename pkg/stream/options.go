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

package stream

import (
	"github.com/spf13/cobra"
	"time"
)

type StreamOptions struct {
	BackendStorageKind string

	// Settings for aws storage
	AwsS3Host      string
	AwsS3AccessKey string
	AwsS3SecretKey string
	AwsS3Region    string
	AwsS3Bucket    string

	// Settings for local storage
	LocalPath string

	//
	CoreFileSurvivalTime        time.Duration
	CoreFileMaxSizePerContainer string
}

var (
	streamOpts               *StreamOptions
	BackendPathSetFunc       func(*cobra.Command) error
	CoreFileSurvivalTimeFunc func(*cobra.Command) error
)

func init() {
	streamOpts = &StreamOptions{}
	BackendPathSetFunc = func(cmd *cobra.Command) error {
		flags := cmd.Flags()
		flags.StringVar(&streamOpts.BackendStorageKind, "backend-kind", streamOpts.BackendStorageKind, "choose which kind of backend, 'asw' or 'local' to use")
		flags.StringVar(&streamOpts.LocalPath, "local-path", streamOpts.LocalPath, "local storage path, only useable when backend-kind is 'local'")

		flags.StringVar(&streamOpts.AwsS3Host, "aws-host", streamOpts.AwsS3Bucket, "aws host, only useable when backend-kind is 'aws'")
		//flags.IntVar(&streamOpts.AwsS3Port, "asw-port", streamOpts.AwsS3Port, "aws port, only useable when backend-kind is 'aws'")
		// flags.IntVar(&streamOpts.AwsS3AccessKeyFile, "asw-access-key-file", streamOpts.AwsS3AccessKeyFile, "aws access key file path, only useable when backend-kind is 'aws'")
		// flags.IntVar(&streamOpts.AwsS3SecretKeyFile, "asw-secret-key-file", streamOpts.AwsS3SecretKeyFile, "aws secret key file path, only useable when backend-kind is 'aws'")
		flags.StringVar(&streamOpts.AwsS3AccessKey, "aws-access-key", streamOpts.AwsS3AccessKey, "aws access key file path, only useable when backend-kind is 'aws'")
		flags.StringVar(&streamOpts.AwsS3SecretKey, "aws-secret-key", streamOpts.AwsS3SecretKey, "aws secret key file path, only useable when backend-kind is 'aws'")
		flags.StringVar(&streamOpts.AwsS3Region, "aws-region", "default", "aws region, only useable when backend-kind is 'aws'")

		flags.StringVar(&streamOpts.AwsS3Bucket, "aws-bucket", streamOpts.AwsS3Bucket, "aws bucket name, only useable when backend-kind is 'aws'")
		return nil
	}

	CoreFileSurvivalTimeFunc = func(cmd *cobra.Command) error {
		flags := cmd.Flags()
		flags.DurationVar(&streamOpts.CoreFileSurvivalTime, "core-file-survival-time", 24*time.Hour, "core file survival time before gc")
		flags.StringVar(&streamOpts.CoreFileMaxSizePerContainer, "core-file-max-per-container", "1GB", "core file max size per container, must be integer, support 'KB', 'MB', 'GB'")
		return nil
	}
}
