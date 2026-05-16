package markdown

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestGoReadabilityExtractorExtractsReadableHTML(t *testing.T) {
	extractor := NewGoReadabilityExtractor()

	output, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.NoError(t, err)
	require.Equal(t, ProviderGoReadability, extractor.Provider())
	require.Equal(t, "Readable Article", output.Title)
	require.Contains(t, output.Markdown, "This paragraph contains enough natural language")
}

func TestGoReadabilityExtractorReturnsLocalUnreadable(t *testing.T) {
	extractor := NewGoReadabilityExtractor()

	output, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><nav>home</nav><button>save</button></body></html>`),
		CanonicalURL: "https://example.com/navigation",
	})

	require.ErrorIs(t, err, arc.ErrLocalUnreadable)
	require.Equal(t, ProviderGoReadability, extractor.Provider())
	require.Empty(t, output.Markdown)
}

func TestGoReadabilityExtractorMapsParseFailureToARC009(t *testing.T) {
	extractor := &GoReadabilityExtractor{
		parseDocument: func(_ []byte) (*html.Node, error) {
			return nil, errors.New("parse failed")
		},
		convert: convertArticleNode,
	}

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.ErrorIs(t, err, arc.ErrLocalExtractionFailed)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, ProviderGoReadability, extractionErr.Provider)
	require.Contains(t, extractionErr.Reason, "parse HTML snapshot")
}

func TestGoReadabilityExtractorMapsConversionFailureToARC009(t *testing.T) {
	extractor := &GoReadabilityExtractor{
		parseDocument: parseHTMLDocument,
		convert: func(_ context.Context, _ *html.Node, _ *url.URL) (string, error) {
			return "", errors.New("conversion failed")
		},
	}

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.ErrorIs(t, err, arc.ErrLocalExtractionFailed)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, ProviderGoReadability, extractionErr.Provider)
	require.Contains(t, extractionErr.Reason, "convert readable HTML to Markdown")
}

func TestGoReadabilityExtractorRejectsInvalidCanonicalURL(t *testing.T) {
	extractor := NewGoReadabilityExtractor()

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: ":bad-url",
	})

	require.ErrorIs(t, err, arc.ErrLocalExtractionFailed)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.NotEmpty(t, extractionErr.Reason)
}

func TestGoReadabilityExtractorRejectsEmptyMarkdown(t *testing.T) {
	extractor := &GoReadabilityExtractor{
		parseDocument: parseHTMLDocument,
		convert: func(_ context.Context, _ *html.Node, _ *url.URL) (string, error) {
			return "", nil
		},
	}

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(readableHTML),
		CanonicalURL: "https://example.com/articles/readable",
	})

	require.ErrorIs(t, err, arc.ErrLocalExtractionFailed)
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
