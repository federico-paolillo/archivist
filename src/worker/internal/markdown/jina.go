package markdown

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/imroc/req/v3"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
)

const (
	jinaReaderBaseURL = "https://r.jina.ai/"
)

type JinaExtractor struct {
	apiKey     string
	baseURL    string
	httpClient *req.Client
}

var _ MarkdownExtractor = (*JinaExtractor)(nil)

func NewJinaExtractor(client *req.Client, apiKey string) *JinaExtractor {
	return &JinaExtractor{
		apiKey:     apiKey,
		baseURL:    jinaReaderBaseURL,
		httpClient: client,
	}
}

func (e *JinaExtractor) Provider() Provider {
	return ProviderJina
}

func (e *JinaExtractor) ExtractMarkdown(ctx context.Context, input ExtractInput) (ExtractOutput, error) {
	markdown, err := e.fetch(ctx, input.CanonicalURL)
	if err != nil {
		return ExtractOutput{}, err
	}

	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ExtractOutput{}, jinaFailure(arc.ErrJinaReaderFailure, "jina reader returned empty markdown", 0)
	}

	return ExtractOutput{
		Markdown: markdown,
	}, nil
}

func (e *JinaExtractor) fetch(ctx context.Context, canonicalURL string) (string, error) {
	requestURL := e.baseURL + canonicalURL

	r := e.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain")

	if e.apiKey != "" {
		r = r.SetBearerAuthToken(e.apiKey)
	}

	resp, err := r.Get(requestURL)
	if err != nil {
		return "", jinaFailure(arc.ErrJinaReaderFailure, fmt.Sprintf("jina reader request failed: %v", err), 0)
	}

	if resp.StatusCode == http.StatusPaymentRequired {
		return "", jinaFailure(
			arc.ErrJinaInsufficientBalance,
			"jina reader insufficient balance",
			resp.StatusCode,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return "", jinaFailure(arc.ErrJinaReaderFailure, "jina reader returned unexpected HTTP status", resp.StatusCode)
	}

	return resp.String(), nil
}
