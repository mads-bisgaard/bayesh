package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3" // imports as package "cli"

	bayesh "github.com/mads-bisgaard/bayesh/src"
	_ "github.com/mattn/go-sqlite3"
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
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(logHandler))
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "settings",
				Usage: "Print settings to stdout",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					settings, err := bayesh.Setup(ctx, osFS{})
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
			{
				Name:  "infer-cmd",
				Usage: "Infer command",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "cwd",
					},
					&cli.StringArg{
						Name: "previous-cmd",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					settings, err := bayesh.Setup(ctx, osFS{})
					if err != nil {
						return err
					}
					previous_cmd := bayesh.ProcessCmd(osFS{}, cmd.StringArg("previous-cmd"))
					cwd := cmd.StringArg("cwd")
					db, err := sql.Open("sqlite3", settings.DB)
					if err != nil {
						return err
					}
					defer func() {
						if err := db.Close(); err != nil {
							log.Fatal("Failed to close DB:", err)
						}
					}()
					queries := bayesh.New(db)
					inferredCmd, err := queries.InferCurrentCmd(ctx, cwd, previous_cmd)
					if err != nil {
						return err
					}
					fmt.Println(strings.Join(inferredCmd, "\n"))
					return nil
				},
			},
		},
	}
	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}

}
