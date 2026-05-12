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

// ErrLocalUnreadable is the ARC-coded public error for locally unreadable documents.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrLocalUnreadable = errors.New("[ARC-008] Archivist could not read this page locally.")

// ErrLocalMarkdownExtraction is the ARC-coded public error for local extraction or conversion failures.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrLocalMarkdownExtraction = errors.New("[ARC-009] Archivist could not extract this page locally.")

// ErrJinaReaderFailure is the ARC-coded public error for Jina Reader fallback failures.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrJinaReaderFailure = errors.New("[ARC-010] Archivist could not extract this page with the fallback reader.")

// ErrJinaInsufficientBalance is the ARC-coded public error for Jina Reader billing failures.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrJinaInsufficientBalance = errors.New("[ARC-011] Archivist could not use the fallback reader because the Jina account is out of credit.")

// ErrMarkdownWrite is the ARC-coded public error for Markdown artifact write failures.
//
//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrMarkdownWrite = errors.New("[ARC-012] Archivist could not store the Markdown article.")

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
