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
		Resolves []string `short:"r"`
		Message  string   `long:"m" short:"m"`
		Amend    bool
		Trailer  bool     `default:"true" negatable:""`
		Args     []string `arg:"" help:"" optional:"" passthrough:"all"`
	}
)

const (
	messageGenTrailerName = "Message-generated-by"
	commitHelp            = `Usage: git do commit [flags] [<rest> ...]

Flags:
  -h, --help 
    Show this help message.
  -r=RESOLVES,..., --resolves=RESOLVES,...
    Issue or ticket identifiers that are resolved by the content of this commit.
    This flag may be included more than once or as a comma-separated list.
  --amend
    Re-generate the most recent commit's message and amend that commit.
  --[no-]trailer
    Include, or omit, the 'Message-generated-by' commit trailer (it will be included by default).
  -m
    A message that will be included in the commit generation prompt.
    This message may be used to alter, inform or fully override the default system prompt.
    
All input provided after the above set of flags will be piped directly to the 'git commit' CLI.
`
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

	//msgVal, foundMsg, msgIndices := git.ExtractFlag(recv.Args, "-m")

	commitMsg, err := ctx.LLM.GenerateCommit(
		ctx, seq,
		llm.CommitWithResolutions(recv.Resolves...),
		llm.CommitWithInstructions(recv.Message),
	)
	if err != nil {
		return err
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
	_, err := dst.Write([]byte(commitHelp))
	return err
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
		llm.CommitWithInstructions(recv.Message),
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
