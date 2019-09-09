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
	backendtypes "github.com/WanLinghao/coredump-detector/pkg/backend/types"
	"github.com/spf13/cobra"
	//	"time"
)

var (
	BackendOpts        *backendtypes.BackendOptions
	BackendPathSetFunc func(*cobra.Command) error
)

func init() {
	BackendOpts = &backendtypes.BackendOptions{}
	BackendPathSetFunc = func(cmd *cobra.Command) error {
		flags := cmd.Flags()
		flags.StringVar(&BackendOpts.BackendStorageKind, "backend-kind", "local", "choose which kind of backend, 'asw' or 'local' to use")
		flags.StringVar(&BackendOpts.LocalPath, "local-path", BackendOpts.LocalPath, "local storage path, only useable when backend-kind is 'local'")

		flags.StringVar(&BackendOpts.AwsS3Host, "aws-host", BackendOpts.AwsS3Bucket, "aws host, only useable when backend-kind is 'aws'")
		//flags.IntVar(&BackendOpts.AwsS3Port, "asw-port", BackendOpts.AwsS3Port, "aws port, only useable when backend-kind is 'aws'")
		// flags.IntVar(&BackendOpts.AwsS3AccessKeyFile, "asw-access-key-file", BackendOpts.AwsS3AccessKeyFile, "aws access key file path, only useable when backend-kind is 'aws'")
		// flags.IntVar(&BackendOpts.AwsS3SecretKeyFile, "asw-secret-key-file", BackendOpts.AwsS3SecretKeyFile, "aws secret key file path, only useable when backend-kind is 'aws'")
		flags.StringVar(&BackendOpts.AwsS3AccessKey, "aws-access-key", BackendOpts.AwsS3AccessKey, "aws access key file path, only useable when backend-kind is 'aws'")
		flags.StringVar(&BackendOpts.AwsS3SecretKey, "aws-secret-key", BackendOpts.AwsS3SecretKey, "aws secret key file path, only useable when backend-kind is 'aws'")
		flags.StringVar(&BackendOpts.AwsS3Region, "aws-region", "default", "aws region, only useable when backend-kind is 'aws'")

		flags.StringVar(&BackendOpts.AwsS3Bucket, "aws-bucket", BackendOpts.AwsS3Bucket, "aws bucket name, only useable when backend-kind is 'aws'")
		return nil
	}

	// CoreFileSurvivalTimeFunc = func(cmd *cobra.Command) error {
	// 	flags := cmd.Flags()
	// 	flags.DurationVar(&BackendOpts.CoreFileSurvivalTime, "core-file-survival-time", 24*time.Hour, "core file survival time before gc")
	// 	flags.StringVar(&BackendOpts.CoreFileMaxSizePerContainer, "core-file-max-per-container", "1GB", "core file max size per container, must be integer, support 'KB', 'MB', 'GB'")
	// 	return nil
	// }
}
