package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgapp "codeberg.org/federico-paolillo/archivist/pkg/app"
)

const defaultProcessIdleSleep = time.Second

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

	processed, err := a.SnapshotPipeline.ProcessOne(ctx)
	if err != nil {
		if contextCancelled(ctx) {
			return true, nil
		}

		return true, fmt.Errorf("worker: process one job: %w", err)
	}

	if once {
		return true, nil
	}

	if processed {
		return false, nil
	}

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
