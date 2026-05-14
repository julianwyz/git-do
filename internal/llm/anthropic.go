package llm

import (
	"context"
	"io"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
)

const anthropicBaseMaxTokens = int64(8192)

type anthropicProvider struct {
	client    anthropic.Client
	model     string
	reasoning ReasoningLevel
}

func newAnthropicProvider(cfg *llmConfig) (*anthropicProvider, error) {
	c := anthropic.NewClient(
		anthropicopt.WithAPIKey(cfg.apiKey),
		anthropicopt.WithBaseURL(cfg.apiBase),
		anthropicopt.WithHTTPClient(cfg.http),
	)

	return &anthropicProvider{
		client:    c,
		model:     cfg.model,
		reasoning: cfg.reasoning,
	}, nil
}

func (p *anthropicProvider) generateCommit(ctx context.Context, instructions string, msgs []llmMessage) (string, error) {
	budget := reasoningBudget(p.reasoning)
	if budget > 0 {
		return p.generateCommitBeta(ctx, instructions, msgs, budget)
	}

	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: anthropicBaseMaxTokens,
		System:    []anthropic.TextBlockParam{{Text: instructions}},
		Messages:  buildMessages(msgs),
	})
	if err != nil {
		return "", err
	}

	log.Debug().
		Int64("input_tokens", resp.Usage.InputTokens).
		Int64("output_tokens", resp.Usage.OutputTokens).
		Msg("anthropic generate")

	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	return text, nil
}

func (p *anthropicProvider) generateCommitBeta(ctx context.Context, instructions string, msgs []llmMessage, budget int64) (string, error) {
	resp, err := p.client.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: budget + anthropicBaseMaxTokens,
		System:    []anthropic.BetaTextBlockParam{{Text: instructions}},
		Messages:  buildBetaMessages(msgs),
		Thinking:  anthropic.BetaThinkingConfigParamOfEnabled(budget),
	}, anthropicopt.WithHeaderAdd("anthropic-beta", "interleaved-thinking-2025-05-14"))
	if err != nil {
		return "", err
	}

	log.Debug().
		Int64("input_tokens", resp.Usage.InputTokens).
		Int64("output_tokens", resp.Usage.OutputTokens).
		Msg("anthropic generate (thinking)")

	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	return text, nil
}

func (p *anthropicProvider) streamOutput(ctx context.Context, instructions string, msgs []llmMessage, dst io.Writer) (int64, int64, error) {
	budget := reasoningBudget(p.reasoning)
	if budget > 0 {
		return p.streamOutputBeta(ctx, instructions, msgs, dst, budget)
	}

	stream := p.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: anthropicBaseMaxTokens,
		System:    []anthropic.TextBlockParam{{Text: instructions}},
		Messages:  buildMessages(msgs),
	})
	acc := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		if err := acc.Accumulate(event); err != nil {
			return 0, 0, err
		}
		if e, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if d, ok := e.Delta.AsAny().(anthropic.TextDelta); ok {
				if _, err := dst.Write([]byte(d.Text)); err != nil {
					return 0, 0, err
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return 0, 0, err
	}

	return acc.Usage.InputTokens, acc.Usage.OutputTokens, nil
}

func (p *anthropicProvider) streamOutputBeta(ctx context.Context, instructions string, msgs []llmMessage, dst io.Writer, budget int64) (int64, int64, error) {
	stream := p.client.Beta.Messages.NewStreaming(ctx, anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: budget + anthropicBaseMaxTokens,
		System:    []anthropic.BetaTextBlockParam{{Text: instructions}},
		Messages:  buildBetaMessages(msgs),
		Thinking:  anthropic.BetaThinkingConfigParamOfEnabled(budget),
	}, anthropicopt.WithHeaderAdd("anthropic-beta", "interleaved-thinking-2025-05-14"))
	acc := anthropic.BetaMessage{}
	for stream.Next() {
		event := stream.Current()
		if err := acc.Accumulate(event); err != nil {
			return 0, 0, err
		}
		if e, ok := event.AsAny().(anthropic.BetaRawContentBlockDeltaEvent); ok {
			if d, ok := e.Delta.AsAny().(anthropic.BetaTextDelta); ok {
				if _, err := dst.Write([]byte(d.Text)); err != nil {
					return 0, 0, err
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return 0, 0, err
	}

	return acc.Usage.InputTokens, acc.Usage.OutputTokens, nil
}

func buildMessages(msgs []llmMessage) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, len(msgs))
	for i, m := range msgs {
		result[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(m.text))
	}

	return result
}

func buildBetaMessages(msgs []llmMessage) []anthropic.BetaMessageParam {
	result := make([]anthropic.BetaMessageParam, len(msgs))
	for i, m := range msgs {
		result[i] = anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(m.text))
	}

	return result
}
