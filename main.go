package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3" // imports as package "cli"

	bayesh "github.com/mads-bisgaard/bayesh/src"
)

//go:embed shell/bayesh.zsh
var bayeshZsh string

//go:embed shell/fzf_tmux_server.zsh
var fzfTmuxServerZsh string

//go:embed shell/capture.zsh
var captureZsh string

// version is set at build time using ldflags.
var version = "development"

func main() {
	ctx := context.Background()
	settings, err := bayesh.Setup(ctx, bayesh.OsFs{})
	if err != nil {
		slog.Error("Error: Failed during initial setup", "error", err)
		os.Exit(1)
	}

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

	var zsh bool

	cmd := &cli.Command{
		Name:    "bayesh",
		Usage:   "CLI for integrating Bayesh into your shell",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "zsh",
				Usage:       "Print the zsh shell integration script",
				Destination: &zsh,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if zsh {
				fmt.Print(fzfTmuxServerZsh)
				fmt.Print(captureZsh)
				fmt.Print(bayeshZsh)
			} else {
				if err := cli.ShowAppHelp(cmd); err != nil {
					return err
				}
			}
			return nil
		},
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
				Name: "infer-cmd",
				// commands are passed as typed: they may be empty, have surrounding
				// whitespace or start with "-", which the flag parser would mangle.
				// So --print0 is only recognized before the cwd, which is absolute.
				SkipFlagParsing: true,
				Usage:           "Infer command; --print0 terminates suggestions with NUL instead of newline, for fzf --read0",
				ArgsUsage:       "[--print0] <cwd> <previous-cmd>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args, sep := cmd.Args().Slice(), "\n"
					if len(args) > 0 && args[0] == "--print0" {
						args, sep = args[1:], "\x00"
					}
					args = append(args, "", "") // missing arguments are empty
					inferredCommands, err := core.InferCommands(ctx, args[0], args[1])
					if err != nil {
						return err
					}
					fmt.Print(bayesh.FormatSuggestions(inferredCommands, sep))
					return nil
				},
			},
			{
				Name: "record-event",
				// commands are passed as typed: they may be empty, have surrounding
				// whitespace or start with "-", which the flag parser would mangle
				SkipFlagParsing: true,
				Usage:           "Record a command event",
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
