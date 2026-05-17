package runner

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
)

func runManyPrograms(
	ctx context.Context,
	logger *slog.Logger,
	appRoot *app.App,
	cfg *config.Root,
	programs ...ProgramE,
) StatusCode {
	programsCount := len(programs)

	// Allow cancellation via SIGTERM

	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalChan)

	// We buffer for n-programs so that each program can report a failure without blocking.

	errChan := make(chan error, programsCount)

	var wg sync.WaitGroup

	for _, p := range programs {
		wg.Go(func() {
			runOneProgram(ctx, errChan, appRoot, cfg, p)
		})
	}

	err := waitForTermination(
		cancel,
		errChan,
		signalChan,
		&wg,
		programsCount,
	)
	if err != nil {
		logger.Error(
			"runner: failed to run a program",
			slog.Any("error", err),
		)

		return NotOk
	}

	return Ok
}

func waitForTermination(
	cancel context.CancelFunc,
	errChan chan error,
	signalChan chan os.Signal,
	wg *sync.WaitGroup,
	programsCount int,
) error {
	var firstErr error

	remainingResults := programsCount

	// As soon as one program fails, finishes, or the context is cancelled we give up execution

	select {
	case firstErr = <-errChan:
		remainingResults--

		cancel()
	case <-signalChan:
		cancel()
	}

	// Once we signalled cancellation we wait for all goroutines to clean up before returning

	wg.Wait()

	errs := make([]error, 0, programsCount)
	if firstErr != nil {
		errs = append(errs, firstErr)
	}

	for range remainingResults {
		err := <-errChan
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func runOneProgram(
	ctx context.Context,
	errChan chan<- error,
	appRoot *app.App,
	cfg *config.Root,
	program ProgramE,
) {
	// errChan has one buffer space for each program. We don't have to worry about blocking write
	errChan <- program(ctx, appRoot, cfg)
}
