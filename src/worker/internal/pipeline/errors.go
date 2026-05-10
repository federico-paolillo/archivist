// Package pipeline orchestrates article-processing pipeline stages.
package pipeline

import "errors"

// ErrSnapshotWrite is the ARC-coded public error for snapshot write failures.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrSnapshotWrite = errors.New("[ARC-007] Archivist could not store the HTML snapshot.")

// ErrUnknown is the ARC-coded public error for any unexpected implementation failure.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrUnknown = errors.New("[ARC-999] Archivist could not process the URL.")
