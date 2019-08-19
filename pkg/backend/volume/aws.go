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
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/types"
)

type awsStorage struct {
	Host      string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	PathStyle bool
}

func NewAwsStorage(host, accessKey, secretKey, region, bucket string, pathStyle bool) (types.Storage, error) {
	return &awsStorage{
		Host:      host,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Region:    region,
		PathStyle: pathStyle,
	}, nil
}

func (a *awsStorage) GetCoreFiles(ns, podUID, container string) (string, error) {
	fileKey := ns + "-" + podUID + "-" + container

	dest := "/tmp/" + fileKey
	err := a.download(fileKey, dest)
	return dest, err
}

func (a *awsStorage) CleanCoreFiles(ns, podUID, container string) error {
	// TODO: implements clean logic
	return nil
}

func (a *awsStorage) CleanNamespace(namespace string) error {
	// TODO: implements clean logic
	return nil
}

func (a *awsStorage) LogPodDeletion(namespace string, podUID string, deletionTimestamp time.Time) error {
	// TODO: implements log pod deletion logic
	return nil
}

func (a *awsStorage) GC() error {
	// TODO: implements gc logic
	return nil
}

// ----------------------------------------------------------
// implements github.com/aws/aws-sdk-go/aws/credentials.Provider interface
func (a *awsStorage) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     a.AccessKey,
		SecretAccessKey: a.SecretKey,
	}, nil
}

func (m *awsStorage) IsExpired() bool { return false }

//----------------------------------------------------------

func (a *awsStorage) connectAws() (*s3.S3, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(a.Region),
			//EndpointResolver: endpoints.ResolverFunc(s3CustResolverFn),
			Endpoint:         &a.Host,
			S3ForcePathStyle: &a.PathStyle,
			Credentials:      credentials.NewCredentials(a),
		},
	}))
	// Create the S3 service client with the shared session. This will
	// automatically use the S3 custom endpoint configured in the custom
	// endpoint resolver wrapping the default endpoint resolver.
	return s3.New(sess), nil
}

func (a *awsStorage) download(fileKey, dst string) error {
	s3Svc, _ := a.connectAws()
	// Operation calls will be made to the custom endpoint.
	resp, err := s3Svc.GetObject(&s3.GetObjectInput{
		Bucket: &a.Bucket,
		Key:    &fileKey,
	})
	if err != nil {
		return fmt.Errorf("get object for %s failed: %v", fileKey, err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0666)
	if out == nil {
		return fmt.Errorf("open file %s failed: %v", dst, err)
	}
	_, err = io.Copy(out, resp.Body)
	return err
}
