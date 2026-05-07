package markdown

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestGoReadabilityExtractorExtractsReadableHTML(t *testing.T) {
	extractor := NewGoReadabilityExtractor()

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.Equal(t, ResultStatusSuccess, result.Status)
	require.Equal(t, ProviderGoReadability, result.Provider)
	require.Empty(t, result.ErrorCode)
	require.Equal(t, "Readable Article", result.Title)
	require.Contains(t, result.Markdown, "This paragraph contains enough natural language")
}

func TestGoReadabilityExtractorReturnsLocalUnreadable(t *testing.T) {
	extractor := NewGoReadabilityExtractor()

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><nav>home</nav><button>save</button></body></html>`),
		CanonicalURL: "https://example.com/navigation",
	})

	require.Equal(t, ResultStatusLocalUnreadable, result.Status)
	require.Equal(t, ProviderGoReadability, result.Provider)
	require.Empty(t, result.Markdown)
	require.Empty(t, result.ErrorCode)
}

func TestGoReadabilityExtractorMapsParseFailureToARC009(t *testing.T) {
	extractor := &GoReadabilityExtractor{
		parseDocument: func(_ []byte) (*html.Node, error) {
			return nil, errors.New("parse failed")
		},
		convert: convertArticleNode,
	}

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.Equal(t, ResultStatusFailure, result.Status)
	require.Equal(t, ProviderGoReadability, result.Provider)
	require.Equal(t, ErrorCodeLocalExtractionFailed, result.ErrorCode)
	require.Contains(t, result.FailureReason, "parse HTML snapshot")
}

func TestGoReadabilityExtractorMapsConversionFailureToARC009(t *testing.T) {
	extractor := &GoReadabilityExtractor{
		parseDocument: parseHTMLDocument,
		convert: func(_ context.Context, _ *html.Node, _ *url.URL) (string, error) {
			return "", errors.New("conversion failed")
		},
	}

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.Equal(t, ResultStatusFailure, result.Status)
	require.Equal(t, ProviderGoReadability, result.Provider)
	require.Equal(t, ErrorCodeLocalExtractionFailed, result.ErrorCode)
	require.Contains(t, result.FailureReason, "convert readable HTML to Markdown")
}

const readableHTML = `<!doctype html>
<html>
<head>
	<title>Readable Article</title>
</head>
<body>
	<article>
		<h1>Readable Article</h1>
		<p>This paragraph contains enough natural language for the readability gate to treat the document as a real article. It describes the local extraction path with concrete prose and avoids navigation-only content.</p>
		<p>The second paragraph adds additional readable body copy so the article has meaningful density. Local extraction should preserve these sentences when converting the readable HTML to Markdown.</p>
	</article>
</body>
</html>`
