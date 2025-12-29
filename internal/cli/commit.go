package cli

import (
	"bytes"

	"github.com/julianwyz/git-buddy/internal/git"
)

type (
	Commit struct {
		Args []string `arg:"" optional:"" passthrough:"all"`
	}
)

func (recv *Commit) Run(ctx *Ctx) error {
	wd := ctx.Value(ctxWD).(string)

	seq, err := git.ListStaged(
		ctx, wd,
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
		wd,
		bytes.NewBufferString(commitMsg),
		recv.Args...,
	)
}
