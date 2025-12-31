package llm

type (
	llmConfig struct {
		apiBase string
		apiKey  string
		model   string
	}

	LLMOpt func(*llmConfig) error
)

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
