package markdown

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
)

type documentParser func([]byte) (*html.Node, error)

type markdownConverter func(context.Context, *html.Node, *url.URL) (string, error)

type GoReadabilityExtractor struct {
	parseDocument documentParser
	convert       markdownConverter
}

var _ MarkdownExtractor = (*GoReadabilityExtractor)(nil)

func NewGoReadabilityExtractor() *GoReadabilityExtractor {
	return &GoReadabilityExtractor{
		parseDocument: parseHTMLDocument,
		convert:       convertArticleNode,
	}
}

func (e *GoReadabilityExtractor) Provider() Provider {
	return ProviderGoReadability
}

func (e *GoReadabilityExtractor) ExtractMarkdown(ctx context.Context, input ExtractInput) (ExtractOutput, error) {
	pageURL, err := url.ParseRequestURI(input.CanonicalURL)
	if err != nil {
		return ExtractOutput{}, localFailure(arc.ErrLocalExtractionFailed, fmt.Sprintf("parse canonical URL: %v", err))
	}

	parseDocument := e.parseDocument
	if parseDocument == nil {
		parseDocument = parseHTMLDocument
	}

	doc, err := parseDocument(input.HTML)
	if err != nil {
		return ExtractOutput{}, localFailure(arc.ErrLocalExtractionFailed, fmt.Sprintf("parse HTML snapshot: %v", err))
	}

	if !readability.CheckDocument(doc) {
		return ExtractOutput{}, localFailure(arc.ErrLocalUnreadable, "local unreadable")
	}

	article, err := readability.FromDocument(doc, pageURL)
	if err != nil {
		return ExtractOutput{}, localFailure(arc.ErrLocalExtractionFailed, fmt.Sprintf("extract readable HTML: %v", err))
	}

	if article.Node == nil {
		return ExtractOutput{}, localFailure(arc.ErrLocalExtractionFailed, "extract readable HTML: empty article node")
	}

	convert := e.convert
	if convert == nil {
		convert = convertArticleNode
	}

	markdown, err := convert(ctx, article.Node, pageURL)
	if err != nil {
		return ExtractOutput{}, localFailure(
			arc.ErrLocalExtractionFailed,
			fmt.Sprintf("convert readable HTML to Markdown: %v", err),
		)
	}

	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ExtractOutput{}, localFailure(arc.ErrLocalExtractionFailed, "convert readable HTML to Markdown: empty Markdown")
	}

	return ExtractOutput{
		Markdown: markdown,
		Title:    article.Title(),
	}, nil
}

func parseHTMLDocument(input []byte) (*html.Node, error) {
	doc, err := html.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("html parse failed: %w", err)
	}

	return doc, nil
}

func convertArticleNode(ctx context.Context, articleNode *html.Node, pageURL *url.URL) (string, error) {
	markdownBytes, err := htmltomarkdown.ConvertNode(
		articleNode,
		converter.WithContext(ctx),
		converter.WithDomain(pageURL.String()),
	)
	if err != nil {
		return "", fmt.Errorf("html-to-markdown conversion failed: %w", err)
	}

	return string(markdownBytes), nil
}
