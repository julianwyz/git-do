package cli

import (
	"bytes"

	"github.com/julianwyz/git-do/internal/git"
)

type (
	Commit struct {
		Args []string `arg:"" help:"" optional:"" passthrough:"all"`
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

func (recv *Commit) Help() string {
	return "hello help"
}
