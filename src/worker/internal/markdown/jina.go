package markdown

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/imroc/req/v3"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
)

const (
	jinaReaderBaseURL        = "https://r.jina.ai/"
	maxJinaMarkdownBytes     = 10 * 1024 * 1024
	maxJinaDiagnosticBytes   = 64 * 1024
	jinaReadLimitExceededMsg = "jina reader response body exceeds size limit"
)

var acceptedJinaContentTypes = []string{"text/plain", "text/markdown", "text/x-markdown"}

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
		SetHeader("Accept", "text/plain").
		DisableAutoReadResponse()

	if e.apiKey != "" {
		r = r.SetBearerAuthToken(e.apiKey)
	}

	resp, err := r.Get(requestURL)
	if err != nil {
		return "", jinaFailure(arc.ErrJinaReaderFailure, fmt.Sprintf("jina reader request failed: %v", err), 0)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := readJinaBody(resp.Body, maxJinaDiagnosticBytes)
		if readErr != nil {
			return "", jinaFailure(arc.ErrJinaReaderFailure, readErr.Error(), resp.StatusCode)
		}

		return "", classifyJinaHTTPError(resp.StatusCode, body)
	}

	if !isAcceptedJinaContentType(resp.GetContentType()) {
		return "", jinaFailure(arc.ErrJinaReaderFailure, "jina reader returned unexpected content type", resp.StatusCode)
	}

	body, readErr := readJinaBody(resp.Body, maxJinaMarkdownBytes)
	if readErr != nil {
		return "", jinaFailure(arc.ErrJinaReaderFailure, readErr.Error(), resp.StatusCode)
	}

	return string(body), nil
}

func isAcceptedJinaContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	for _, accepted := range acceptedJinaContentTypes {
		if strings.EqualFold(mediaType, accepted) {
			return true
		}
	}

	return false
}

func readJinaBody(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read jina reader response body: %w", err)
	}

	if int64(len(data)) > maxBytes {
		return nil, errors.New(jinaReadLimitExceededMsg)
	}

	return data, nil
}
