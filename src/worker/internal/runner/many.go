package runner

import (
	"context"
	"log/slog"
	"os"
	"time"

	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
)

/*
 * Channels aren't like files; you don't usually need to close them.
 * Closing is only necessary when the receiver must be told there are no more values coming.
 * We take advantage of this fact to share errChan among all program goroutines.
 * Even though a program goroutine should own the output channel (to close it) we make the main goroutin has owner.
 * There is no need for a program to tell main that no new errors will come. The channel is effectively shared.
 * As soon as one program terminates all programs will terminate. They share the same lifespan
 */

func RunMany(
	ctx context.Context,
	programs ...ProgramE,
) StatusCode {
	logLevel := new(slog.LevelVar)

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}),
	)

	cfg, err := config.Load()
	if err != nil {
		logger.Error(
			"runner: failed to load configuration",
			slog.Any("err", err),
		)

		return NotOk
	}

	if cfg.Debug {
		logLevel.Set(slog.LevelDebug)
	}

	appRoot, err := app.NewApp(ctx, logger, logLevel, cfg)
	if err != nil {
		logger.Error(
			"runner: failed to create app",
			slog.Any("err", err),
		)

		return NotOk
	}

	defer func() {
		err := appRoot.Close()
		if err != nil {
			logger.Error(
				"runner: failed to close app",
				slog.Any("err", err),
			)
		}
	}()

	startTime := time.Now()

	statusCode := runManyPrograms(ctx, logger, appRoot, cfg, programs...)

	runtimeDuration := time.Since(startTime)

	logger.Info(
		"runner: program has completed. This does not indicate success",
		slog.Float64("runtime_s", runtimeDuration.Seconds()),
		slog.Any("status_code", statusCode),
	)

	return statusCode
}
