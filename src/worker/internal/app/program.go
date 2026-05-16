package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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

func CliProgram(ctx context.Context, a *pkgapp.App, cfg *config.Root) error {
	cmd := &cli.Command{
		Name:  "archivist-worker",
		Usage: "Injests web-pages sent to the Archivist Telegram bot",
		Flags: globalFlags(),
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			applyDebugFlag(cmd, cfg, a)

			return ctx, nil
		},
		Commands: []*cli.Command{
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
					return process(ctx, a, cmd.Bool("once"), cmd.Duration("idle-sleep"))
				},
			},
			{
				Name:  "version",
				Usage: "Dummy command. Prints the Go version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return version(ctx, a)
				},
			},
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	return nil
}