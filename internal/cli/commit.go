package cli

import (
	"bytes"
	"context"
	"io"
	"slices"

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
	if recv.Amend {
		return recv.amendCommit(ctx)
	}

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

func (recv *Commit) amendCommit(ctx *Ctx) error {
	headRef, err := git.HeadHash(ctx, ctx.WorkingDir)
	if err != nil {
		return err
	}

	seq, err := git.ListCommitChanges(
		ctx, ctx.WorkingDir,
		headRef,
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

	args := slices.Concat(
		recv.Args,
		[]string{"--amend"},
	)

	return git.Commit(
		ctx,
		ctx.WorkingDir,
		bytes.NewBufferString(commitMsg),
		args...,
	)
}
