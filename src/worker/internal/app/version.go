package app

import (
	"context"
	"log/slog"
	"runtime"

	pkgapp "codeberg.org/federico-paolillo/archivist/pkg/app"
)

func version(ctx context.Context, a *pkgapp.App) error {
	version := runtime.Version()

	a.Logger.InfoContext(ctx, "version: version detected successfully",
		slog.String("go_version", version),
	)

	return nil
}