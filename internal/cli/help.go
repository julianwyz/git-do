package cli

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/glamour"
)

type (
	helper interface {
		Help(dst io.Writer) error
	}
)

func (recv *CLI) OutputHelp(to io.Writer) kong.HelpPrinter {
	return func(options kong.HelpOptions, cli *kong.Context) error {
		var (
			handler any
			cmd     = cli.Command()
		)

		switch cmd {
		case "commit":
			handler = recv.Commit
		case "explain":
			handler = recv.Explain
		case "status":
			handler = recv.Status
		}

		if handler != nil {
			if h, ok := handler.(helper); ok {
				return h.Help(to)
			}
		}

		_, err := fmt.Fprintf(to, "No help documentation available for '%s' command.\n", cmd)
		return err
	}
}

func renderHelpMarkdown(dst io.Writer, content string) error {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	if err != nil {
		return err
	}

	s, err := r.Render(content)
	if err != nil {
		return err
	}

	_, err = dst.Write([]byte(s))
	return err
}
