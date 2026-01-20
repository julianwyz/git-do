package cli

import (
	"io"
	"os"

	"github.com/julianwyz/git-do/internal/git"
)

type (
	Status struct {
		Pathspec string `arg:"" optional:""`
	}
)

const (
	statusHelp = `git do status [pathspec]
=======

Flags:

` + "`-h`" + `, ` + "`--help`" + `
> Show this help message.

Arguments:

` + "`[pathspec]`" + `
> The pathspec to check the status of (defaults to ` + "`.`" + `)
`
)

func (recv *Status) Run(ctx *Ctx) error {
	pathspec := "."
	if len(recv.Pathspec) > 0 {
		pathspec = recv.Pathspec
	}

	seq, status, err := git.Status(
		ctx, ctx.WorkingDir, pathspec,
	)
	if err != nil {
		return err
	}

	err = ctx.LLM.ExplainStatus(
		ctx, status, seq,
		os.Stdout,
	)
	return err
}

func (recv Status) Help(dst io.Writer) error {
	return renderHelpMarkdown(dst, statusHelp)
}
