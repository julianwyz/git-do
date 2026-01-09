package cli

import (
	"io"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/julianwyz/git-do/internal/git"
)

type (
	Explain struct {
		From string `arg:"" optional:""`
		To   string `arg:"" optional:""`
	}
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

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return err
	}

	if err := ctx.LLM.ExplainCommits(
		ctx, commitIter,
		renderer,
	); err != nil {
		return err
	}
	if err := renderer.Close(); err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, renderer)
	return err
}
