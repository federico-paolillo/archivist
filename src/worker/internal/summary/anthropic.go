package summary

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
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
	logger *slog.Logger
}

var _ SummarizerService = (*AnthropicAdapter)(nil)

// NewAnthropicAdapter constructs an AnthropicAdapter from configuration values.
// apiKey must not be logged.
func NewAnthropicAdapter(apiKey, model string, logger *slog.Logger) *AnthropicAdapter {
	return newAdapter(apiKey, model, "", logger)
}

// NewAnthropicAdapterWithBaseURL constructs an AnthropicAdapter that sends requests
// to a custom base URL. Intended for testing with httptest servers.
func NewAnthropicAdapterWithBaseURL(apiKey, model, baseURL string, logger *slog.Logger) *AnthropicAdapter {
	return newAdapter(apiKey, model, baseURL, logger)
}

func newAdapter(apiKey, model, baseURL string, logger *slog.Logger) *AnthropicAdapter {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(opts...)

	return &AnthropicAdapter{
		client: client,
		model:  model,
		logger: logger,
	}
}

// Summarize sends the Markdown source to Claude and returns a text summary.
// It classifies Anthropic errors into ARC-013, ARC-014, and ARC-015.
func (a *AnthropicAdapter) Summarize(ctx context.Context, req SummarizerRequest) SummarizerResult {
	msg, err := a.client.Messages.New(ctx, a.buildParams(req))
	if err != nil {
		return a.classifyError(err)
	}

	text := a.extractText(msg)
	if text == "" {
		a.logger.Error("summary: anthropic returned empty text output", "model", a.model)

		return SummarizerResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderAnthropic,
			ErrorCode:     ErrorCodeProviderFailure,
			FailureReason: "provider returned empty text output",
		}
	}

	a.logger.Info(
		"summary: anthropic summarization succeeded",
		"model", a.model,
		"request_id", msg.ID,
	)

	return SummarizerResult{
		Status:   ResultStatusSuccess,
		Provider: ProviderAnthropic,
		Summary:  text,
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

func (a *AnthropicAdapter) classifyError(err error) SummarizerResult {
	apiErr, ok := errors.AsType[*anthropic.Error](err)
	if ok {
		return a.classifyAPIError(apiErr)
	}

	a.logger.Error("summary: anthropic transport or unknown error", "model", a.model, "error", err)

	return SummarizerResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderAnthropic,
		ErrorCode:     ErrorCodeProviderFailure,
		FailureReason: fmt.Sprintf("provider error: %v", err),
	}
}

func (a *AnthropicAdapter) classifyAPIError(apiErr *anthropic.Error) SummarizerResult {
	errType := apiErr.Type()

	a.logger.Error(
		"summary: anthropic API error",
		"model", a.model,
		"status_code", apiErr.StatusCode,
		"error_type", string(errType),
		"request_id", apiErr.RequestID,
	)

	if isBillingError(apiErr) {
		return SummarizerResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderAnthropic,
			ErrorCode:     ErrorCodeBillingFailure,
			FailureReason: fmt.Sprintf("billing error (HTTP %d): %v", apiErr.StatusCode, errType),
		}
	}

	if isTooLargeError(apiErr) {
		return SummarizerResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderAnthropic,
			ErrorCode:     ErrorCodeRequestTooLarge,
			FailureReason: fmt.Sprintf("request too large (HTTP %d): %v", apiErr.StatusCode, errType),
		}
	}

	return SummarizerResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderAnthropic,
		ErrorCode:     ErrorCodeProviderFailure,
		FailureReason: fmt.Sprintf("provider error (HTTP %d): %v", apiErr.StatusCode, errType),
	}
}

func isBillingError(apiErr *anthropic.Error) bool {
	return apiErr.StatusCode == http.StatusPaymentRequired || apiErr.Type() == anthropic.ErrorTypeBillingError
}

func isTooLargeError(apiErr *anthropic.Error) bool {
	return apiErr.StatusCode == http.StatusRequestEntityTooLarge
}
