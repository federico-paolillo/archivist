package markdown

import "context"

type Provider string

const (
	ProviderGoReadability Provider = "go-readability"
)

type ExtractInput struct {
	HTML         []byte
	CanonicalURL string
}

type ExtractOutput struct {
	Markdown string
	Title    string
}

type MarkdownExtractor interface {
	Provider() Provider
	ExtractMarkdown(ctx context.Context, input ExtractInput) (ExtractOutput, error)
}
