package cli

import "github.com/julianwyz/git-buddy/internal/config"

type (
	cliConfig struct {
		userConfig *config.Config
	}

	CLIOpt func(*cliConfig) error
)

func WithConfig(c *config.Config) CLIOpt {
	return func(cc *cliConfig) error {
		cc.userConfig = c

		return nil
	}
}
