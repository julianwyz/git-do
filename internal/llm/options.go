package llm

import "github.com/julianwyz/git-do/internal/config"

type (
	llmConfig struct {
		apiBase string
		apiKey  string
		model   string
		config  *config.Config
	}

	LLMOpt func(*llmConfig) error
)

func WithConfig(c *config.Config) LLMOpt {
	return func(lc *llmConfig) error {
		lc.config = c

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
