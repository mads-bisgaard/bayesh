package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3" // imports as package "cli"

	bayesh "github.com/mads-bisgaard/bayesh/src"
)

func main() {
	ctx := context.Background()
	settings, err := bayesh.Setup(ctx, bayesh.OsFs{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed during initial setup: %v\n", err)
		os.Exit(1)
	}
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: settings.LogLevel})
	slog.SetDefault(slog.New(logHandler))
	core, err := bayesh.NewCore(ctx, settings)
	if err != nil {
		slog.Error("Failed to create core", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := core.Close(); err != nil {
			slog.Error("Failed to close core", "error", err)
			os.Exit(1)
		}
	}()

	cmd := &cli.Command{
		Name:  "bayesh",
		Usage: "CLI for integrating Bayesh into your shell",
		Commands: []*cli.Command{
			{
				Name:  "settings",
				Usage: "Print settings to stdout",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					jsonSettings, err := core.Settings.ToJSON()
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
					previous_cmd := cmd.StringArg("previous-cmd")
					cwd := cmd.StringArg("cwd")
					inferredCommands, err := core.InferCommands(ctx, cwd, previous_cmd)
					if err != nil {
						return err
					}
					fmt.Println(bayesh.AnsiColorTokens(strings.Join(inferredCommands, "\n")))
					return nil
				},
			},
			{
				Name:  "record-event",
				Usage: "Record a command event",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "cwd",
					},
					&cli.StringArg{
						Name: "previous-cmd",
					},
					&cli.StringArg{
						Name: "current-cmd",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cwd := cmd.StringArg("cwd")
					previousCmd := cmd.StringArg("previous-cmd")
					currentCmd := cmd.StringArg("current-cmd")
					return core.RecordEvent(ctx, cwd, previousCmd, currentCmd)
				},
			},
		},
	}
	if err := cmd.Run(ctx, os.Args); err != nil {
		slog.Error("CLI command failed", "error", err)
		os.Exit(1)
	}

}
