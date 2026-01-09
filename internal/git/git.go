package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type (
	CommitFormat string
)

const (
	CommitFormatGithub       = CommitFormat("github")
	CommitFormatConventional = CommitFormat("conventional")
)

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

func HelpOf(
	ctx context.Context,
	wd,
	cmd string,
	dst io.Writer,
) error {
	return prepareGitCmd(
		ctx,
		wd,
		dst,
		dst,
		cmd,
		"--help",
	).Run()
}

func DiffsOfCommit(
	ctx context.Context,
	wd,
	ref,
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
		"--raw",
		fmt.Sprintf("%s^", ref),
		ref,
		"--",
		target,
	).Run()
}

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
		target,
	).Run()
}

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

	return cmd
}
