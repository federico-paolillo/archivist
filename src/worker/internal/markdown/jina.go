package markdown

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/imroc/req/v3"
)

const (
	jinaReaderBaseURL = "https://r.jina.ai/"
)

type JinaExtractor struct {
	enabled    bool
	apiKey     string
	baseURL    string
	httpClient *req.Client
}

var _ MarkdownExtractor = (*JinaExtractor)(nil)

func NewJinaExtractor(client *req.Client, enabled bool, apiKey string) *JinaExtractor {
	return &JinaExtractor{
		enabled:    enabled,
		apiKey:     apiKey,
		baseURL:    jinaReaderBaseURL,
		httpClient: client,
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

func (e *JinaExtractor) fetch(ctx context.Context, canonicalURL string) (string, *jinaError) {
	requestURL := e.baseURL + canonicalURL

	r := e.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain")

	if e.apiKey != "" {
		r = r.SetBearerAuthToken(e.apiKey)
	}

	resp, err := r.Get(requestURL)
	if err != nil {
		return "", &jinaError{
			code:   ErrorCodeJinaFailed,
			reason: fmt.Sprintf("jina reader request failed: %v", err),
		}
	}

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

	return resp.String(), nil
}

