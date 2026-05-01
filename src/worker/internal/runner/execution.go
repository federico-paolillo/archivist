package runner

import (
	"context"
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

	// We buffer for n-programs. So that each program can report a failure without blocking
	// TODO: We currently consume just the first program error

	errChan := make(chan error, programsCount)

	var wg sync.WaitGroup

	wg.Add(programsCount)

	for _, p := range programs {
		go runOneProgram(
			ctx,
			&wg,
			errChan,
			appRoot,
			cfg,
			p,
		)
	}

	err := waitForTermination(
		cancel,
		errChan,
		signalChan,
		&wg,
	)
	if err != nil {
		logger.Error(
			"runner: failed to run a program",
			slog.Any("err", err),
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
) error {
	var err error

	// As soon as one program fails, finishes, or the context is cancelled we give up execution
	// TODO: We should consume all errors. The channel has a finite number of values

	select {
	case err = <-errChan:
		cancel()
	case <-signalChan:
		cancel()
	}

	// Once we signalled cancellation we wait for all goroutines to clean up before returning

	wg.Wait()

	return err
}

func runOneProgram(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	appRoot *app.App,
	cfg *config.Root,
	program ProgramE,
) {
	defer wg.Done()

	// errChan has one buffer space for each program. We don't have to worry about blocking write

	errChan <- program(ctx, appRoot, cfg)
}
