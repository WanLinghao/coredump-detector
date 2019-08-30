package main

import "fmt"
import "bytes"
import "encoding/base64"
import "log"
import "io/ioutil"

func main() {
	file := "/home/wlh/mywork/go/packages/src/github.com/WanLinghao/fujitsu-coredump/config/certificates/apiserver_ca.crt"
	buff := bytes.Buffer{}
	enc := base64.NewEncoder(base64.StdEncoding, &buff)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Could not read file %s: %v", file, err)
	}

	_, err = enc.Write(data)
	if err != nil {
		log.Fatalf("Could not write bytes: %v", err)
	}
	enc.Close()
	fmt.Printf("%s", buff.String())
}
