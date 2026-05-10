// Package pipeline orchestrates article-processing pipeline stages.
package pipeline

import (
	"errors"
	"strings"
)

// ErrSnapshotWrite is the ARC-coded public error for snapshot write failures.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrSnapshotWrite = errors.New("[ARC-007] Archivist could not store the HTML snapshot.")

// ErrUnknown is the ARC-coded public error for any unexpected implementation failure.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrUnknown = errors.New("[ARC-999] Archivist could not process the URL.")

// isARCError returns true when err carries an ARC-coded message prefix ("[ARC-").
func isARCError(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "[ARC-")
}

// arcCode extracts the ARC-NNN token from an ARC-coded error, or returns "ARC-999".
func arcCode(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	// Expected format: "[ARC-NNN] ..." — closing bracket at index 8.
	if strings.HasPrefix(msg, "[ARC-") && len(msg) >= 9 && msg[8] == ']' {
		return msg[1:8]
	}

	return "ARC-999"
}
