package arc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatReturnsPublicMessage(t *testing.T) {
	require.Equal(t, "[ARC-003] The URL was not found.", Format(CodeURLNotFound))
}

func TestUnknownFormatFallsBackToARC999(t *testing.T) {
	require.Equal(t, "[ARC-999] Archivist could not process the URL.", Format(Code("ARC-777")))
}

func TestErrorsIsMatchesByCode(t *testing.T) {
	err := fmt.Errorf("fetcher: %w", ErrURLNotFound)

	require.ErrorIs(t, err, ErrURLNotFound)
	require.NotErrorIs(t, err, ErrURLAccessDenied)
}

func TestCodeOfExtractsWrappedARCError(t *testing.T) {
	err := fmt.Errorf("anthropic: request failed: %w", ErrSummarizerProviderFailure)

	code, ok := CodeOf(err)
	require.True(t, ok)
	require.Equal(t, CodeSummarizerProviderFailed, code)
	require.Equal(t, "ARC-013", CodeString(err))
}

func TestPublicMessageIgnoresWrappedDiagnosticContext(t *testing.T) {
	err := fmt.Errorf("anthropic: HTTP 500: %w", ErrSummarizerProviderFailure)

	message, ok := PublicMessage(err)
	require.True(t, ok)
	require.Equal(t, "[ARC-013] Archivist could not summarize this article.", message)
	require.Contains(t, err.Error(), "HTTP 500")
}

func TestCodeOfSupportsErrorsAs(t *testing.T) {
	err := fmt.Errorf("markdown: %w", ErrMarkdownWrite)

	arcErr, ok := errors.AsType[*Error](err)
	require.True(t, ok)
	require.Equal(t, CodeMarkdownWriteFailed, arcErr.Code())
}

func TestCodeStringMapsUnknownNonNilErrorToARC999(t *testing.T) {
	require.Equal(t, "ARC-999", CodeString(errors.New("plain error")))
	require.Empty(t, CodeString(nil))
}
