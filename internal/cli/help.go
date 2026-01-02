package cli

import (
	"context"
	"io"

	"github.com/alecthomas/kong"
)

func (recv *CLI) OutputHelp(to io.Writer) kong.HelpPrinter {
	return func(options kong.HelpOptions, cli *kong.Context) error {
		ctx := context.TODO()

		switch cli.Command() {
		case "commit":
			return recv.Commit.Help(ctx, to)
		}

		return nil
	}
}
