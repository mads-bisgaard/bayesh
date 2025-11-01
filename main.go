package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3" // imports as package "cli"

	bayesh "github.com/mads-bisgaard/bayesh/src"
)

type osFS struct{}

func (osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (osFS) Create(name string) (*os.File, error) {
	return os.Create(name)
}
func (osFS) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
func (osFS) Getenv(key string) string {
	return os.Getenv(key)
}
func (osFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func main() {
	ctx := context.Background()
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "settings",
				Usage: "Print settings to stdout",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					settings, err := bayesh.Initialize(ctx, osFS{})
					if err != nil {
						return err
					}
					jsonSettings, err := settings.ToJSON()
					if err != nil {
						return err
					}
					fmt.Println(jsonSettings)
					return nil
				},
			},
		},
	}
	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}

}
