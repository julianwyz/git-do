package cli

type (
	cliConfig struct {
	}

	CLIOpt func(*cliConfig) error
)
