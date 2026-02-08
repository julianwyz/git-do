package cli

type (
	cliConfig struct {
		output destination
		input  destination
		wd     string
		hd     string
	}

	CLIOpt func(*cliConfig) error
)

func WithOutput(d destination) CLIOpt {
	return func(cc *cliConfig) error {
		cc.output = d

		return nil
	}
}

func WithInput(d destination) CLIOpt {
	return func(cc *cliConfig) error {
		cc.input = d

		return nil
	}
}

func WithWorkingDir(d string) CLIOpt {
	return func(cc *cliConfig) error {
		cc.wd = d

		return nil
	}
}

func WithHomeDir(d string) CLIOpt {
	return func(cc *cliConfig) error {
		cc.hd = d

		return nil
	}
}
