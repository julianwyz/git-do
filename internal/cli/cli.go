package cli

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
)

type (
	CLI struct {
		runner *kong.Context `kong:"-"`

		Commit Commit `cmd:"" help:"commit"`
	}

	Ctx struct {
		context.Context
	}

	ctxKey string
)

const (
	ctxWD = ctxKey("working_directory")
)

func New() *CLI {
	returner := &CLI{}
	returner.runner = kong.Parse(returner)

	return returner
}

func (recv *CLI) Exec(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, ctxWD, cwd)

	return recv.runner.Run(&Ctx{
		Context: ctx,
	})
}

func (recv *CLI) FatalIfErrorf(err error, args ...any) {
	recv.runner.FatalIfErrorf(err, args...)
}
