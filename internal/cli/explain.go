package cli

import (
	"io"

	"github.com/charmbracelet/glamour"
	"github.com/julianwyz/git-do/internal/git"
)

type (
	Explain struct {
		From  string `arg:"" optional:""`
		To    string `arg:"" optional:""`
		Plain bool   `optional:""`
	}
)

const (
	explainHelp = `git do explain [ref] [to-ref]
=======

Flags:

` + "`-h`" + `, ` + "`--help`" + `
> Show this help message.

` + "`--plain`" + `
> Output the explanation without markdown rendering.

Arguments:

` + "`[ref]`" + `
> The commit reference to explain. If omitted, ` + "`HEAD`" + ` is used.

` + "`[to-ref]`" + `
> The lower-bound commit reference to explain.
>
> All commits between ` + "`[ref]`" + `and ` + "`[to-ref]`" + ` will be included in the explanation.
>
> If omitted, only the ` + "`[ref]`" + ` is explained.
`
)

func (recv *Explain) Run(ctx *Ctx) error {
	commitRefs := make([]string, 0, 2)
	if len(recv.From) > 0 {
		commitRefs = append(commitRefs, recv.From)
		if len(recv.To) > 0 {
			commitRefs = append(commitRefs, recv.To)
		}
	} else {
		commitRefs = append(commitRefs, "HEAD")
	}

	commitIter, err := git.CommitsBetween(
		ctx,
		ctx.WorkingDir,
		commitRefs,
	)
	if err != nil {
		return err
	}

	var (
		outputDst io.ReadWriteCloser = ctx.Output
		finalize  func() error
	)
	if !recv.Plain && !ctx.PipedOutput {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
		)
		if err != nil {
			return err
		}

		outputDst = renderer

		finalize = func() error {
			if err := outputDst.Close(); err != nil {
				return err
			}

			_, err = io.Copy(ctx.Output, outputDst)

			return err
		}
	}

	if err := ctx.LLM.ExplainCommits(
		ctx, commitIter,
		outputDst,
	); err != nil {
		return err
	}

	if finalize != nil {
		return finalize()
	}

	return nil
}

func (recv Explain) Help(dst io.Writer) error {
	return renderHelpMarkdown(dst, explainHelp)
}
