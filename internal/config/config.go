package config

import (
	"github.com/BurntSushi/toml"
)

type (
	Config struct {
		LLM *LLM `toml:"llm"`
	}

	LLM struct {
		APIBase string `toml:"api_base"`
		APIKey  string `toml:"api_key"`
		Model   string `toml:"model"`
	}
)

func Load(fp string) (*Config, error) {
	var dst Config
	if _, err := toml.DecodeFile(fp, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}
