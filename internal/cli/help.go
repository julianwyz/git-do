package cli

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"sort"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/glamour"
)

type (
	Help struct{}

	helper interface {
		Help(dst io.Writer) error
	}
)

var (
	helpMap = map[string]helper{
		"init":    Init{},
		"status":  Status{},
		"explain": Explain{},
		"commit":  Commit{},
	}
)

func (recv *Help) Run(ctx *Ctx) error {
	return recv.Help(os.Stdout)
}

func (recv Help) Help(dst io.Writer) error {
	keys := slices.Collect(maps.Keys(helpMap))
	sort.Strings(keys)
	for _, s := range keys {
		if err := helpOf(dst, s); err != nil {
			return err
		}
	}

	return nil
}

func (recv *CLI) OutputHelp(to io.Writer) kong.HelpPrinter {
	return func(options kong.HelpOptions, cli *kong.Context) error {
		return helpOf(to, cli.Command())
	}
}

func helpOf(to io.Writer, cmd string) error {
	hlpr, f := helpMap[cmd]
	if !f {
		_, err := fmt.Fprintf(to, "No help documentation available for '%s' command.\n", cmd)

		return err
	}

	return hlpr.Help(to)
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
