package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	pkgapp "codeberg.org/federico-paolillo/archivist/pkg/app"
)

const defaultProcessIdleSleep = 300 * time.Second

func process(ctx context.Context, a *pkgapp.App, once bool, idleSleep time.Duration) error {
	if idleSleep <= 0 {
		return errors.New("worker: idle sleep must be positive")
	}

	for {
		exit, err := runProcessLoopIteration(ctx, a, once, idleSleep)
		if err != nil {
			return err
		}

		if exit {
			return nil
		}
	}
}

func runProcessLoopIteration(
	ctx context.Context,
	a *pkgapp.App,
	once bool,
	idleSleep time.Duration,
) (bool, error) {
	if contextCancelled(ctx) {
		return true, nil
	}

	a.Logger.Info(
		"worker: process loop iteration started",
		slog.String("stage", "process_loop"),
		slog.String("status", "start"),
		slog.Bool("once", once),
	)

	processed, err := a.SnapshotPipeline.ProcessOne(ctx)
	if err != nil {
		if contextCancelled(ctx) {
			return true, nil
		}

		return true, fmt.Errorf("worker: process one job: %w", err)
	}

	if once {
		if !processed {
			a.Logger.Info(
				"worker: process poll result",
				slog.String("stage", "process_loop"),
				slog.String("status", "idle"),
				slog.Bool("processed", false),
			)
		}

		return true, nil
	}

	if processed {
		return false, nil
	}

	a.Logger.Info(
		"worker: process poll result",
		slog.String("stage", "process_loop"),
		slog.String("status", "idle"),
		slog.Bool("processed", false),
		slog.Duration("idle_sleep", idleSleep),
	)

	return waitForNextProcessPoll(ctx, idleSleep), nil
}

func waitForNextProcessPoll(ctx context.Context, idleSleep time.Duration) bool {
	timer := time.NewTimer(idleSleep)

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		return true
	case <-timer.C:
		return false
	}
}

func contextCancelled(ctx context.Context) bool {
	return ctx.Err() != nil
}
