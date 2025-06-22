package main

import (
	"fmt"
	"os"

	bayesh "github.com/mads-bisgaard/bayesh/src"
)

func main() {
	rslt := bayesh.ProcessCmd(osFS{}, "echo ./myfile.txt")
	fmt.Println(rslt)
}

type osFS struct{}

func (osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
