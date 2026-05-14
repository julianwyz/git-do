package llm

import (
	"context"
	"io"
)

type provider interface {
	generateCommit(ctx context.Context, instructions string, msgs []llmMessage) (string, error)
	streamOutput(ctx context.Context, instructions string, msgs []llmMessage, dst io.Writer) (tokensIn, tokensOut int64, err error)
}

type llmMessage struct {
	text string
}

func reasoningBudget(l ReasoningLevel) int64 {
	switch l {
	case ReasoningLevelMinimal:
		return 1024
	case ReasoningLevelLow:
		return 2048
	case ReasoningLevelMedium:
		return 4096
	case ReasoningLevelHigh:
		return 8192
	case ReasoningLevelXHigh:
		return 16384
	default:
		return 0
	}
}
