package git

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type (
	CommitFormat string
)

const (
	CommitFormatGithub       = CommitFormat("github")
	CommitFormatConventional = CommitFormat("conventional")
)

// Init a git repo.
func Init(ctx context.Context, wd string, out io.Writer) error {
	return prepareGitCmd(
		ctx,
		wd,
		out,
		out,
		"init",
	).Run()
}

// Status of the target pathspec within the git repo located at the provided wd.
func Status(
	ctx context.Context,
	wd string,
	target string,
) (iter.Seq2[string, error], string, error) {
	var (
		affectedFiles = &bytes.Buffer{}
		statusOut     = &bytes.Buffer{}
		porcelainErr  error
		regErr        error
		wg            sync.WaitGroup
	)

	wg.Go(func() {
		porcelainErr = prepareGitCmd(
			ctx,
			wd,
			affectedFiles,
			os.Stderr,
			"status",
			"--porcelain",
			target,
		).Run()
	})

	wg.Go(func() {
		regErr = prepareGitCmd(
			ctx,
			wd,
			statusOut,
			os.Stderr,
			"status",
			target,
		).Run()
	})

	wg.Wait()

	if porcelainErr != nil || regErr != nil {
		return nil, "", errors.Join(porcelainErr, regErr)
	}

	scanner := bufio.NewScanner(affectedFiles)

	return func(yield func(string, error) bool) {
		for scanner.Scan() {
			txt := scanner.Text()
			if len(txt) < 2 {
				continue
			}

			flag := txt[0:2]
			fp := strings.TrimSpace(txt[2:])
			diffFlags := []string{}
			diffPaths := []string{"--"}
			// change flag is a two-character string.
			// first char represents the staged change
			// second is the unstaged change
			// the position of the flag tells us where the
			// change is in the pipeline
			switch flag {
			case "??",
				" A":
				// new file or untracked.
				// diff is the entire file
				diffFlags = append(diffFlags, "--no-index")
				diffPaths = append(diffPaths, os.DevNull, fp)
			case "R ":
				// rename will be "before -> after"
				{
					s := strings.Split(fp, "->")

					diffFlags = append(diffFlags, "--staged")
					diffPaths = append(diffPaths, strings.TrimSpace(s[len(s)-1]))
				}
			case "M ", "A ", "D ":
				// staged and...
				// modified, added
				diffFlags = append(diffFlags, "--staged")
				diffPaths = append(diffPaths, fp)
			default:
				// known file - can diff directly
				diffPaths = append(diffPaths, fp)
			}

			buf := &bytes.Buffer{}
			err := prepareGitCmd(
				ctx,
				wd,
				buf,
				os.Stderr,
				slices.Concat([]string{"diff"}, diffFlags, diffPaths)...,
			).Run()
			if err != nil && strings.Contains(err.Error(), "exit status 1") {
				// when using no-index and dev/null, git will use a 1 error code
				// we ignore this
				err = nil
			}

			if !yield(buf.String(), err) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield("", err)

			return
		}
	}, statusOut.String(), nil
}

// HeadHash of the git repo at wd.
func HeadHash(ctx context.Context, wd string) (string, error) {
	buf := &bytes.Buffer{}

	if err := prepareGitCmd(
		ctx,
		wd,
		buf,
		nil,
		"rev-parse",
		"HEAD",
	).Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// CommitsBetween the provided refRange
//
// This provides an iterator to step through the commits
// in the range.
//
// refRange is expected to have...
//
//	[0] = start
//	[1] = end
//
// If a single ref is provided, that will be the only commit
// returned by the iterator.
func CommitsBetween(
	ctx context.Context,
	wd string,
	refRange []string,
) (iter.Seq2[string, error], error) {
	if len(refRange) == 1 {
		refRange = append(refRange, refRange[0])
	}

	ref := strings.Join(refRange, "..")
	commitBatch := &bytes.Buffer{}
	if err := prepareGitCmd(
		ctx,
		wd,
		commitBatch,
		os.Stderr,
		"log",
		"--pretty=format:%H",
		ref,
	).Run(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(commitBatch)

	return func(yield func(string, error) bool) {
		doHash := func(hash string) bool {
			log.Debug().
				Str("hash", hash).
				Msg("include commit in summary")

			buf := &bytes.Buffer{}
			err := ShowCommit(ctx, wd, hash, buf)

			return yield(buf.String(), err)
		}

		for scanner.Scan() {
			if !doHash(scanner.Text()) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield("", err)

			return
		}

		// do our base commit as it won't be included
		// in the list
		if !doHash(refRange[0]) {
			return
		}
	}, nil
}

// ShowCommit identified by ref at the git repo at wd
//
// stdout will be piped to the dst.
func ShowCommit(
	ctx context.Context,
	wd,
	ref string,
	dst io.Writer,
) error {
	return prepareGitCmd(
		ctx,
		wd,
		dst,
		os.Stderr,
		"show",
		ref,
	).Run()
}

// DiffsOfCommit writes the git patch
// of changes made to the target pathspec at the provided ref.
func DiffsOfCommit(
	ctx context.Context,
	wd,
	ref,
	target string,
	dst io.Writer,
) error {
	cmtArgs := []string{
		fmt.Sprintf("%s^", ref),
		ref,
	}

	rc, _ := RootCommit(ctx, wd)
	if rc == ref {
		dn, err := hashDevNull(ctx, wd)
		if err != nil {
			return err
		}
		// we are on the root commit, do not use the parent
		cmtArgs = []string{
			dn,
			ref,
		}
	}

	cmdArgs := slices.Concat(
		[]string{
			"diff",
			"--unified=12",
			"--raw",
		},
		cmtArgs,
		[]string{
			"--",
			target,
		},
	)

	return prepareGitCmd(
		ctx,
		wd,
		dst,
		os.Stderr,
		cmdArgs...,
	).Run()
}

// RootCommit hash of the git repo at wd.
func RootCommit(
	ctx context.Context,
	wd string,
) (string, error) {
	var dst bytes.Buffer
	if err := prepareGitCmd(
		ctx,
		wd,
		&dst,
		os.Stderr,
		"rev-list",
		"--max-parents=0",
		"HEAD",
	).Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(
		dst.String(),
	), nil
}

// StagedDiffs writes the staged diffs of target
// to the dst.
func StagedDiffs(
	ctx context.Context,
	wd,
	target string,
	dst io.Writer,
) error {
	return prepareGitCmd(
		ctx,
		wd,
		dst,
		os.Stderr,
		"diff",
		"--unified=12",
		"--cached",
		"--raw",
		"--",
		target,
	).Run()
}

// ListCommitChanges provides an iterator to step through
// the changes to each file committed at ref.
func ListCommitChanges(ctx context.Context, wd, ref string) (iter.Seq2[string, error], error) {
	fileList := &bytes.Buffer{}
	if err := prepareGitCmd(
		ctx,
		wd,
		fileList,
		os.Stderr,
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
		ref,
	).Run(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fileList)

	return func(yield func(string, error) bool) {
		for scanner.Scan() {
			fileName := scanner.Text()
			diffs := &bytes.Buffer{}
			err := DiffsOfCommit(
				ctx,
				wd,
				ref,
				fileName,
				diffs,
			)

			if !yield(diffs.String(), err) {
				return
			}
		}
	}, nil
}

// ListStaged provides an iterator to step through each file
// that currently has staged changes.
func ListStaged(ctx context.Context, wd string) (iter.Seq2[string, error], error) {
	fileList := &bytes.Buffer{}
	if err := prepareGitCmd(
		ctx,
		wd,
		fileList,
		os.Stderr,
		"diff",
		"--name-only",
		"--cached",
	).Run(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fileList)

	return func(yield func(string, error) bool) {
		for scanner.Scan() {
			fileName := scanner.Text()
			diffs := &bytes.Buffer{}
			err := StagedDiffs(ctx, wd, fileName, diffs)

			if !yield(diffs.String(), err) {
				return
			}
		}
	}, nil
}

// Commit the changes to the local git repo.
func Commit(
	ctx context.Context,
	wd string,
	msg io.Reader,
	args ...string,
) error {
	cmdLine := slices.Concat(
		[]string{"commit"},
		args,
	)

	if msg != nil {
		// take from stdin
		cmdLine = append(cmdLine, "-F", "-")
	}

	cmd := prepareGitCmd(
		ctx,
		wd,
		os.Stdout,
		os.Stderr,
		cmdLine...,
	)

	cmd.Stdin = msg

	return cmd.Run()
}

func hashDevNull(ctx context.Context, wd string) (string, error) {
	var dst bytes.Buffer
	if err := prepareGitCmd(
		ctx,
		wd,
		&dst,
		os.Stderr,
		"hash-object",
		"-t",
		"tree",
		os.DevNull,
	).Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(
		dst.String(),
	), nil
}

func prepareGitCmd(
	ctx context.Context,
	wd string,
	stdout io.Writer,
	stderr io.Writer,
	args ...string,
) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		"git", args...,
	)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = wd
	log.Debug().
		Stringer("cmd", cmd).
		Str("wd", wd).
		Msg("git")

	return cmd
}
