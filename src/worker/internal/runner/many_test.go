package runner_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/runner"
	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/stretchr/testify/require"
)

func TestRunnerReturnsNotOkWhenProgramFails(t *testing.T) {
	ctx := t.Context()
	setRequiredWorkerConfig(t)

	testProgram := func(
		_ context.Context,
		_ *app.App,
		_ *config.Root,
	) error {
		return errors.New("failure is expected")
	}

	statusCode := runner.RunMany(ctx, testProgram)

	require.Equal(t, runner.NotOk, statusCode)
}

func TestRunnerReturnsOkWhenProgramSucceeds(t *testing.T) {
	ctx := t.Context()
	setRequiredWorkerConfig(t)

	testProgram := func(
		_ context.Context,
		_ *app.App,
		_ *config.Root,
	) error {
		return nil
	}

	statusCode := runner.RunMany(ctx, testProgram)

	require.Equal(t, runner.Ok, statusCode)
}

func TestRunnerReturnsOkWhenAllProgramSucceeds(t *testing.T) {
	ctx := t.Context()
	setRequiredWorkerConfig(t)

	testProgram1 := func(
		_ context.Context,
		_ *app.App,
		_ *config.Root,
	) error {
		return nil
	}

	testProgram2 := func(
		_ context.Context,
		_ *app.App,
		_ *config.Root,
	) error {
		return nil
	}

	statusCode := runner.RunMany(ctx, testProgram1, testProgram2)

	require.Equal(t, runner.Ok, statusCode)
}

func TestRunnerReturnsOkWhenOneProgramFails(t *testing.T) {
	ctx := t.Context()
	setRequiredWorkerConfig(t)

	testProgram1 := func(
		_ context.Context,
		_ *app.App,
		_ *config.Root,
	) error {
		time.Sleep(1 * time.Second) // Simulate work to let the other program fail
		return nil
	}

	testProgram2 := func(
		_ context.Context,
		_ *app.App,
		_ *config.Root,
	) error {
		return errors.New("failure is expected")
	}

	statusCode := runner.RunMany(ctx, testProgram1, testProgram2)

	require.Equal(t, runner.NotOk, statusCode)
}

func setRequiredWorkerConfig(t *testing.T) {
	t.Helper()

	t.Setenv("ARCHIVIST_SQLITE_PATH", filepath.Join(t.TempDir(), "archive.db"))
	t.Setenv("ARCHIVIST_DATA_DIR", t.TempDir())
	t.Setenv("ARCHIVIST_LLM_API_KEY", "llm-secret")
}
