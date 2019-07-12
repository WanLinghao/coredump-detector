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
	"os"
	"path/filepath"
)

type localStorage struct {
	rootPath string
}

func NewLocalStorage(rootPath string) (Storage, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("the root path for local storage can't be empty")
	}
	return &localStorage{rootPath}, nil
}
func (l *localStorage) GetCoreFiles(ns, podUID, container string) (string, error) {
	coreDirPath := filepath.Clean(l.rootPath + "/" + ns + "/" + podUID + "/" + container)
	fmt.Printf("we are tar %s\n", coreDirPath)

	stat, err := os.Stat(coreDirPath)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("the %s/%s has no core files in namespace %s", podUID, container, ns)
		} else {
			return "", fmt.Errorf("got unexpected error when access %s: %v", coreDirPath, err)
		}
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("got unexpected error: %s is a not a directory", coreDirPath)
	}

	return makeTar(coreDirPath, "/tmp")
}
