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

// osFS implements src.FileSystem using the real os.Stat
// This allows you to use src.ProcessCmd from main
// and is idiomatic for dependency injection in Go

type osFS struct{}

func (osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
