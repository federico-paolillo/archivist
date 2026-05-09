package summary

import (
	"context"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/imroc/req/v3"
)

// systemPrompt instructs Claude to produce text-only summaries.
const systemPrompt = "You are a helpful assistant that summarizes articles. " +
	"Produce a concise text-only summary of the article provided by the user. " +
	"Do not use structured formats, JSON, or Markdown. " +
	"Write plain prose only."

// maxSummaryTokens is the maximum number of tokens for the generated summary.
const maxSummaryTokens = 1024

// AnthropicAdapter implements SummarizerService using the official Anthropic Go SDK.
// All Anthropic SDK types are private to this file.
type AnthropicAdapter struct {
	client anthropic.Client
	model  string
}

var _ SummarizerService = (*AnthropicAdapter)(nil)

// NewAnthropicAdapter constructs an AnthropicAdapter from configuration values.
// apiKey must not be logged.
func NewAnthropicAdapter(httpClient *req.Client, apiKey, model string) *AnthropicAdapter {
	return newAdapter(httpClient, apiKey, model, "")
}

// NewAnthropicAdapterWithBaseURL constructs an AnthropicAdapter that sends requests
// to a custom base URL. Intended for testing with httptest servers.
func NewAnthropicAdapterWithBaseURL(httpClient *req.Client, apiKey, model, baseURL string) *AnthropicAdapter {
	return newAdapter(httpClient, apiKey, model, baseURL)
}

func newAdapter(httpClient *req.Client, apiKey, model, baseURL string) *AnthropicAdapter {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient.GetClient()),
	}

	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(opts...)

	return &AnthropicAdapter{
		client: client,
		model:  model,
	}
}

// Summarize sends the Markdown source to Claude and returns a text summary.
// It classifies Anthropic errors into ARC-013, ARC-014, and ARC-015.
func (a *AnthropicAdapter) Summarize(ctx context.Context, req SummarizerRequest) SummarizerResult {
	msg, err := a.client.Messages.New(ctx, a.buildParams(req))
	if err != nil {
		return classifyError(err)
	}

	text := a.extractText(msg)
	if text == "" {
		return SummarizerResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderAnthropic,
			ErrorCode:     ErrorCodeProviderFailure,
			FailureReason: "provider returned empty text output",
			RequestID:     msg.ID,
		}
	}

	return SummarizerResult{
		Status:    ResultStatusSuccess,
		Provider:  ProviderAnthropic,
		Summary:   text,
		RequestID: msg.ID,
	}
}

func (a *AnthropicAdapter) buildParams(req SummarizerRequest) anthropic.MessageNewParams {
	return anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: maxSummaryTokens,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.MarkdownSource)),
		},
	}
}

func (a *AnthropicAdapter) extractText(msg *anthropic.Message) string {
	var sb strings.Builder

	for _, block := range msg.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}

	return strings.TrimSpace(sb.String())
}
