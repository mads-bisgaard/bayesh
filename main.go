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

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "process-cmd",
				Usage: "Process a command and return the result",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "cmd",
						Usage:    "The command to process",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					command := cmd.String("cmd")
					result := bayesh.ProcessCmd(osFS{}, command)
					fmt.Println(result)
					return nil
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
