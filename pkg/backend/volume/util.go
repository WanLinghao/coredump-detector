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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func makeTar(sourceDir, destBase string) (string, error) {
	destFile, err := ioutil.TempFile(destBase, "coredump-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	gw := gzip.NewWriter(destFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	fd, err := os.Open(sourceDir)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	err = compress(fd, "", tw)
	if err != nil {
		return "", err
	}

	return destFile.Name(), nil
}

func compress(file *os.File, prefix string, tw *tar.Writer) error {
	fmt.Printf("the name is %s, %#v the prefix is %s\n", file.Name(), file, prefix)
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
