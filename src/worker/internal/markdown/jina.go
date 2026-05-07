package markdown

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ProviderJina Provider = "jina"

	ErrorCodeJinaFailed             ErrorCode = "ARC-010"
	ErrorCodeJinaInsufficientCredit ErrorCode = "ARC-011"

	jinaReaderBaseURL = "https://r.jina.ai/"
)

type JinaExtractor struct {
	enabled    bool
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

var _ MarkdownExtractor = (*JinaExtractor)(nil)

func NewJinaExtractor(enabled bool, apiKey string) *JinaExtractor {
	return &JinaExtractor{
		enabled:    enabled,
		apiKey:     apiKey,
		baseURL:    jinaReaderBaseURL,
		httpClient: &http.Client{},
	}
}

func (e *JinaExtractor) ExtractMarkdown(ctx context.Context, input ExtractInput) ExtractResult {
	if !e.enabled {
		return ExtractResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderJina,
			ErrorCode:     ErrorCodeJinaFailed,
			FailureReason: "jina extractor is disabled",
		}
	}

	markdown, err := e.fetch(ctx, input.CanonicalURL)
	if err != nil {
		return jinaFailure(err.code, err.reason)
	}

	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return jinaFailure(ErrorCodeJinaFailed, "jina reader returned empty markdown")
	}

	return ExtractResult{
		Status:   ResultStatusSuccess,
		Provider: ProviderJina,
		Markdown: markdown,
	}
}

type jinaError struct {
	code   ErrorCode
	reason string
}

func (e *JinaExtractor) fetch(ctx context.Context, canonicalURL string) (string, *jinaError) {
	requestURL := e.baseURL + canonicalURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return "", &jinaError{
			code:   ErrorCodeJinaFailed,
			reason: fmt.Sprintf("build jina reader request: %v", err),
		}
	}

	req.Header.Set("Accept", "text/plain")

	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", &jinaError{
			code:   ErrorCodeJinaFailed,
			reason: fmt.Sprintf("jina reader request failed: %v", err),
		}
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusPaymentRequired {
		return "", &jinaError{
			code:   ErrorCodeJinaInsufficientCredit,
			reason: fmt.Sprintf("jina reader insufficient balance: HTTP %d", resp.StatusCode),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return "", &jinaError{
			code:   ErrorCodeJinaFailed,
			reason: fmt.Sprintf("jina reader returned HTTP %d", resp.StatusCode),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &jinaError{
			code:   ErrorCodeJinaFailed,
			reason: fmt.Sprintf("read jina reader response: %v", err),
		}
	}

	return string(body), nil
}

func jinaFailure(code ErrorCode, reason string) ExtractResult {
	return ExtractResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderJina,
		ErrorCode:     code,
		FailureReason: reason,
	}
}
