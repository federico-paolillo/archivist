package runner

import (
	"context"

	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
)

// ProgramE is a unit of work managed by the runner.
// Programs receive a context (cancelled on shutdown signals), the composition
// root, and the configuration. The naming follows the Cobra RunE convention.
type ProgramE func(context.Context, *app.App, *config.Root) error

type StatusCode string

const (
	Ok    StatusCode = "ok"
	NotOk StatusCode = "nok"
)
