package llm

import (
	"context"
	"io"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

type openaiProvider struct {
	client    *openai.Client
	model     string
	reasoning ReasoningLevel
}

func newOpenAIProvider(cfg *llmConfig) (*openaiProvider, error) {
	c := openai.NewClient(
		option.WithBaseURL(cfg.apiBase),
		option.WithAPIKey(cfg.apiKey),
		option.WithHTTPClient(cfg.http),
	)
	return &openaiProvider{
		client:    &c,
		model:     cfg.model,
		reasoning: cfg.reasoning,
	}, nil
}

func (p *openaiProvider) generateCommit(ctx context.Context, instructions string, msgs []llmMessage) (string, error) {
	resp, err := p.client.Responses.New(ctx, p.buildParams(instructions, msgs))
	if err != nil {
		return "", err
	}
	return resp.OutputText(), nil
}

func (p *openaiProvider) streamOutput(ctx context.Context, instructions string, msgs []llmMessage, dst io.Writer) (int64, int64, error) {
	var tokensIn, tokensOut int64
	stream := p.client.Responses.NewStreaming(ctx, p.buildParams(instructions, msgs))
	for stream.Next() {
		cur := stream.Current()
		if _, err := dst.Write([]byte(cur.Delta)); err != nil {
			return tokensIn, tokensOut, err
		}
		tokensIn += cur.Response.Usage.InputTokens
		tokensOut += cur.Response.Usage.OutputTokens
	}
	return tokensIn, tokensOut, stream.Err()
}

func (p *openaiProvider) buildParams(instructions string, msgs []llmMessage) responses.ResponseNewParams {
	input := make(responses.ResponseInputParam, 0, len(msgs))
	for _, m := range msgs {
		input = append(input, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: param.NewOpt(m.text),
				},
			},
		})
	}
	rp := responses.ResponseNewParams{
		Model:        p.model,
		Instructions: param.NewOpt(instructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
	}
	if effort := openaiReasoningEffort(p.reasoning); effort != "" {
		rp.Reasoning = shared.ReasoningParam{
			Effort: shared.ReasoningEffort(effort),
		}
	}
	return rp
}

func openaiReasoningEffort(l ReasoningLevel) string {
	switch l {
	case ReasoningLevelMinimal, ReasoningLevelLow:
		return "low"
	case ReasoningLevelMedium:
		return "medium"
	case ReasoningLevelHigh, ReasoningLevelXHigh:
		return "high"
	default:
		return ""
	}
}
