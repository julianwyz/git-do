package cli

import (
	"io"
	"os"

	"github.com/julianwyz/git-do/internal/git"
)

type (
	Status struct {
		Resolves []string `short:"r"`
		Message  []string `short:"m"`
		Amend    bool
		Trailer  bool     `default:"true" negatable:""`
		Args     []string `arg:"" optional:"" passthrough:"all"`
	}
)

const (
	statusHelp = `git do status [flags] [rest...]
=======

Flags:

` + "`-h`" + `, ` + "`--help`" + `
> Show this help message.

---

**All input provided after the above set of flags will be piped directly to the ` + "`git status`" + ` CLI.**
`
)

func (recv *Status) Run(ctx *Ctx) error {
	seq, status, err := git.Status(
		ctx, ctx.WorkingDir, ".",
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
