package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	pkgapp "codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/urfave/cli/v3"
)

func applyDebugFlag(cmd *cli.Command, cfg *config.Root, a *pkgapp.App) {
	if !cmd.Bool("debug") {
		return
	}

	cfg.Debug = true

	a.LogLevel.Set(slog.LevelDebug)
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{Name: "debug", Usage: "Enable debug-level logging"},
	}
}

type processCommandOptions struct {
	Once      bool
	IdleSleep time.Duration
}

func runEnqueueCommand(ctx context.Context, a *pkgapp.App, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("enqueue: expected exactly one URL argument, got %d", len(args))
	}

	return enqueue(ctx, a, args[0])
}

func runProcessCommand(ctx context.Context, a *pkgapp.App, opts processCommandOptions) error {
	return process(ctx, a, opts.Once, opts.IdleSleep)
}

func runVersionCommand(ctx context.Context, a *pkgapp.App) error {
	return version(ctx, a)
}

func newCLICommand(a *pkgapp.App, cfg *config.Root) *cli.Command {
	return &cli.Command{
		Name:  "archivist-worker",
		Usage: "Injests web-pages sent to the Archivist Telegram bot",
		Flags: globalFlags(),
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			applyDebugFlag(cmd, cfg, a)

			return ctx, nil
		},
		Commands: []*cli.Command{
			{
				Name:      "enqueue",
				Usage:     "Enqueue an article URL for processing",
				ArgsUsage: "<url>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runEnqueueCommand(ctx, a, cmd.Args().Slice())
				},
			},
			{
				Name:  "process",
				Usage: "Process queued article jobs",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "once",
						Usage: "Process at most one queued job and exit",
					},
					&cli.DurationFlag{
						Name:  "idle-sleep",
						Usage: "Sleep duration when no queued job is available",
						Value: defaultProcessIdleSleep,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runProcessCommand(ctx, a, processCommandOptions{
						Once:      cmd.Bool("once"),
						IdleSleep: cmd.Duration("idle-sleep"),
					})
				},
			},
			{
				Name:  "version",
				Usage: "Dummy command. Prints the Go version",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return runVersionCommand(ctx, a)
				},
			},
		},
	}
}

func runCLIProgram(ctx context.Context, a *pkgapp.App, cfg *config.Root, args []string) error {
	cmd := newCLICommand(a, cfg)

	err := cmd.Run(ctx, args)
	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	return nil
}

func CliProgram(ctx context.Context, a *pkgapp.App, cfg *config.Root) error {
	return runCLIProgram(ctx, a, cfg, os.Args)
}
