package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/julianwyz/git-buddy/internal/git"
	"github.com/simonfrey/jsonl"
)

type (
	Commit struct {
		Args []string `arg:"" optional:"" passthrough:"all"`
	}

	commitDiff struct {
		File  string `json:"file"`
		Diffs string `json:"diff"`
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

	patchFile, err := os.CreateTemp("", ".buddy-patch-*.tmp")
	if err != nil {
		return err
	}
	defer patchFile.Close()
	fmt.Println(patchFile.Name())
	jsonWriter := jsonl.NewWriter(patchFile)

	for staged := range seq {
		fmt.Println("STAGED", staged)

		if err := recv.appendDiffs(
			ctx,
			&jsonWriter,
			wd,
			staged,
		); err != nil {
			return err
		}
	}

	/*git.Commit(
		ctx,
		wd,
		"hello world",
		recv.Args...,
	)*/

	return nil
}

func (recv *Commit) appendDiffs(
	ctx context.Context,
	dst *jsonl.Writer,
	wd,
	file string,
) error {
	diff := &commitDiff{
		File: file,
	}
	w := &bytes.Buffer{}

	if err := git.StagedDiffs(
		ctx,
		wd,
		file,
		w,
	); err != nil {
		return err
	}

	diff.Diffs = w.String()

	return dst.Write(diff)
}
