package cli

import (
	"context"
	"io"

	"github.com/alecthomas/kong"
	"github.com/julianwyz/git-do/internal/git"
)

func (recv *CLI) OutputHelp(to io.Writer) kong.HelpPrinter {
	return func(options kong.HelpOptions, cli *kong.Context) error {
		ctx := context.TODO()

		switch cli.Command() {
		case "commit":
			return git.HelpOf(
				ctx,
				recv.cwd,
				"commit",
				to,
			)
		}

		return nil
	}
}
