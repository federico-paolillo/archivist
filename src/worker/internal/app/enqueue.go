package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	pkgapp "codeberg.org/federico-paolillo/archivist/pkg/app"
)

func enqueue(ctx context.Context, a *pkgapp.App, rawURL string) error {
	normalizedURL, err := validateEnqueueURL(rawURL)
	if err != nil {
		return err
	}

	result, err := a.Enqueuer.EnqueueURL(ctx, normalizedURL)
	if err != nil {
		return fmt.Errorf("worker: enqueue URL: %w", err)
	}

	a.Logger.InfoContext(
		ctx,
		"enqueue: URL queued",
		slog.String("article_id", result.ArticleID),
		slog.String("job_id", result.JobID),
		slog.String("url", normalizedURL),
	)

	return nil
}

func validateEnqueueURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("worker: invalid enqueue URL: %w", err)
	}

	if !parsed.IsAbs() || parsed.Host == "" {
		return "", errors.New("worker: enqueue URL must be an absolute http or https URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("worker: enqueue URL must use http or https")
	}

	return parsed.String(), nil
}
