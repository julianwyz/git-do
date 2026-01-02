package cli

import (
	"bytes"
	"context"
	"io"

	"github.com/julianwyz/git-do/internal/git"
	"github.com/julianwyz/git-do/internal/llm"
)

type (
	Commit struct {
		Resolves []string
		Amend    bool
		Args     []string `arg:"" help:"" optional:"" passthrough:"all"`
	}
)

func (recv *Commit) Run(ctx *Ctx) error {
	seq, err := git.ListStaged(
		ctx, ctx.WorkingDir,
	)
	if err != nil {
		return err
	}

	commitMsg, err := ctx.LLM.GenerateCommit(
		ctx, seq,
		llm.CommitWithResolutions(recv.Resolves...),
	)
	if err != nil {
		return err
	}

	return git.Commit(
		ctx,
		ctx.WorkingDir,
		bytes.NewBufferString(commitMsg),
		recv.Args...,
	)
}

func (recv *Commit) Help(ctx context.Context, dst io.Writer) error {
	return git.HelpOf(
		ctx,
		"",
		"commit",
		dst,
	)
}
