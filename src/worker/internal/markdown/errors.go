package markdown

import (
	"fmt"
	"net/http"
	"strings"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
)

const (
	ProviderJina Provider = "jina"
)

// ExtractionError carries provider diagnostics while unwrapping to an ARC
// sentinel for classification and public persistence.
type ExtractionError struct {
	Provider   Provider
	Reason     string
	StatusCode int
	Err        error
}

func (e *ExtractionError) Error() string {
	if e == nil {
		return ""
	}

	if e.StatusCode != 0 {
		return fmt.Sprintf("markdown: %s: %s (HTTP %d): %v", e.Provider, e.Reason, e.StatusCode, e.Err)
	}

	if e.Reason != "" {
		return fmt.Sprintf("markdown: %s: %s: %v", e.Provider, e.Reason, e.Err)
	}

	return fmt.Sprintf("markdown: %s: %v", e.Provider, e.Err)
}

func (e *ExtractionError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func extractionFailure(provider Provider, err error, reason string, statusCode int) error {
	return &ExtractionError{
		Provider:   provider,
		Reason:     reason,
		StatusCode: statusCode,
		Err:        err,
	}
}

func localFailure(err error, reason string) error {
	return extractionFailure(ProviderGoReadability, err, reason, 0)
}

func jinaFailure(err error, reason string, statusCode int) error {
	return extractionFailure(ProviderJina, err, reason, statusCode)
}

func classifyJinaHTTPError(statusCode int, body []byte) error {
	if statusCode == http.StatusPaymentRequired || containsInsufficientBalanceMarker(body) {
		return jinaFailure(arc.ErrJinaInsufficientBalance, "jina reader insufficient balance", statusCode)
	}

	return jinaFailure(arc.ErrJinaReaderFailure, "jina reader returned unexpected HTTP status", statusCode)
}

func containsInsufficientBalanceMarker(body []byte) bool {
	normalized := normalizeJinaErrorBody(string(body))
	markers := []string{
		"insufficient balance",
		"insufficient credit",
		"insufficient credits",
		"insufficient quota",
		"out of balance",
		"out of credit",
		"out of credits",
		"out of token",
		"out of tokens",
	}

	for _, marker := range markers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	return false
}

func normalizeJinaErrorBody(body string) string {
	body = strings.ToLower(body)
	replacer := strings.NewReplacer(
		"_", " ",
		"-", " ",
		".", " ",
		":", " ",
		"\"", " ",
		"'", " ",
		"{", " ",
		"}", " ",
		"[", " ",
		"]", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)

	return strings.Join(strings.Fields(replacer.Replace(body)), " ")
}
