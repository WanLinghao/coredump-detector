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

package gc

import (
	//"github.com/spf13/cobra"
	//	"time"
	//	"strings"
	//	"strconv"
	"fmt"
)

func convertSizeToByte(size string) (ret int64, err error) {
	// size = strings.ToUpper(size)

	// if strings.HasSuffix(size, "GB") {
	// 	gb, err := strconv.ParseInt(strconv.Atoi(strings.TrimSuffix(size, "GB")), 10, 64)
	// 	if err != nil {
	// 		return ret, err
	// 	}
	// 	return gb * 1024 * 1024 * 1024, nil
	// } else if strings.HasSuffix(size, "MB") {
	// 	mb, err := strconv.ParseInt(strconv.Atoi(strings.TrimSuffix(size, "MB")), 10, 64)
	// 	if err != nil {
	// 		return ret, err
	// 	}
	// 	return mb * 1024 * 1024, nil
	// } else if strings.HasSuffix(size, "KB") {
	// 	kb, err := strconv.ParseInt(strconv.Atoi(strings.TrimSuffix(size, "KB")), 10, 64)
	// 	if err != nil {
	// 		return ret, err
	// 	}
	// 	return kb * 1024, nil
	// } else if strings.HasSuffix(size, "B") {
	// 	b, err := strconv.ParseInt(strconv.Atoi(strings.TrimSuffix(size, "B")), 10, 64)
	// 	if err != nil {
	// 		return ret, err
	// 	}
	// 	return b, nil
	// }
	return ret, fmt.Errorf("error happens when parese limit size")
}
