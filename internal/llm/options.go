package llm

import (
	"github.com/julianwyz/git-do/internal/git"
	"github.com/openai/openai-go/v3/option"
	"golang.org/x/text/language"
)

type (
	llmConfig struct {
		commitFormat  git.CommitFormat
		outputLang    *language.Tag
		apiBase       string
		apiKey        string
		model         string
		reasoning     ReasoningLevel
		contextLoader contextLoader
		http          option.HTTPClient
	}

	commitConfig struct {
		resolutions  []string
		instructions string
	}

	LLMOpt    func(*llmConfig) error
	CommitOpt func(*commitConfig) error
)

func CommitWithInstructions(i string) CommitOpt {
	return func(cc *commitConfig) error {
		cc.instructions = i

		return nil
	}
}

func CommitWithResolutions(rs ...string) CommitOpt {
	return func(cc *commitConfig) error {
		cc.resolutions = append(cc.resolutions, rs...)

		return nil
	}
}

func WithHTTPClient(c option.HTTPClient) LLMOpt {
	return func(lc *llmConfig) error {
		lc.http = c

		return nil
	}
}

func WithOutputLanguage(l language.Tag) LLMOpt {
	return func(lc *llmConfig) error {
		lc.outputLang = &l

		return nil
	}
}

func WithCommitFormat(format git.CommitFormat) LLMOpt {
	return func(lc *llmConfig) error {
		lc.commitFormat = format

		return nil
	}
}

func WithContextLoader(l contextLoader) LLMOpt {
	return func(lc *llmConfig) error {
		lc.contextLoader = l

		return nil
	}
}

func WithAPIBase(base string) LLMOpt {
	return func(lc *llmConfig) error {
		lc.apiBase = base

		return nil
	}
}

func WithAPIKey(key string) LLMOpt {
	return func(lc *llmConfig) error {
		lc.apiKey = key

		return nil
	}
}

func WithModel(m string) LLMOpt {
	return func(lc *llmConfig) error {
		lc.model = m

		return nil
	}
}

func WithReasoningLevel(l ReasoningLevel) LLMOpt {
	return func(lc *llmConfig) error {
		lc.reasoning = l

		return nil
	}
}
