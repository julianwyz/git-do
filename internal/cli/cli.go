package cli

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	"github.com/julianwyz/git-buddy/internal/config"
	"github.com/julianwyz/git-buddy/internal/llm"
)

type (
	CLI struct {
		Commit Commit `cmd:"" help:"commit"`

		runner *kong.Context `kong:"-"`
		config *cliConfig
		llm    *llm.LLM
	}

	Ctx struct {
		context.Context
		LLM        *llm.LLM
		UserConfig *config.Config
	}

	ctxKey string
)

const (
	ctxWD = ctxKey("working_directory")
)

func New(opts ...CLIOpt) (*CLI, error) {
	returner := &CLI{
		config: &cliConfig{},
	}

	for _, o := range opts {
		if err := o(returner.config); err != nil {
			return nil, err
		}
	}

	returner.runner = kong.Parse(returner)

	llm, err := llm.New(returner.config.userConfig.LLM)
	if err != nil {
		return nil, err
	}
	returner.llm = llm

	return returner, nil
}

func (recv *CLI) Exec(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, ctxWD, cwd)

	return recv.runner.Run(&Ctx{
		Context:    ctx,
		LLM:        recv.llm,
		UserConfig: recv.config.userConfig,
	})
}

func (recv *CLI) FatalIfErrorf(err error, args ...any) {
	recv.runner.FatalIfErrorf(err, args...)
}
