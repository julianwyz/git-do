package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/julianwyz/git-do/internal/git"
	"github.com/julianwyz/git-do/internal/llm"
	"github.com/rs/zerolog/log"
)

type (
	Commit struct {
		Resolves []string
		Amend    bool
		Trailer  bool     `default:"true" negatable:""`
		Args     []string `arg:"" help:"" optional:"" passthrough:"all"`
	}
)

const (
	messageGenTrailerName = "Message-generated-by"
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

	msgVal, foundMsg, msgIndices := git.ExtractFlag(recv.Args, "-m")

	commitMsg, err := ctx.LLM.GenerateCommit(
		ctx, seq,
		llm.CommitWithResolutions(recv.Resolves...),
		llm.CommitWithInstructions(msgVal),
	)
	if err != nil {
		return err
	}

	if foundMsg {
		for _, i := range msgIndices {
			recv.Args = slices.Delete(recv.Args, i, i+1)
		}
	}

	if recv.Trailer {
		commitMsg += fmt.Sprintf("\n\n%s",
			recv.commitTrailer(ctx),
		)
	}

	log.Debug().Msgf("commit msg:\n%s", commitMsg)

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

	msgVal, foundMsg, msgIndices := git.ExtractFlag(recv.Args, "-m")

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
		llm.CommitWithInstructions(msgVal),
	)
	if err != nil {
		return err
	}

	if foundMsg {
		for _, i := range msgIndices {
			recv.Args = slices.Delete(recv.Args, i, i+1)
		}
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

func (recv *Commit) commitTrailer(ctx *Ctx) string {
	components := [][]string{
		{
			"git-do",
			Version,
		},
	}

	if ctx != nil && ctx.LLM != nil {
		components = append(components, []string{
			ctx.LLM.GetAPIDomain(), ctx.LLM.GetModel(),
		})
	}

	returner := messageGenTrailerName + ": "
	for _, c := range components {
		returner += strings.Join(c, "/")
		returner += " "
	}

	return strings.TrimSpace(returner)
}
