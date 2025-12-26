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
)

func StagedDiffs(
	ctx context.Context,
	wd,
	target string,
	dst io.Writer,
) error {
	return execGitCmd(
		ctx,
		wd,
		dst,
		os.Stderr,
		"diff",
		"--unified=12",
		"--cached",
		"--raw",
		target,
	)
}

func ListStaged(ctx context.Context, wd string) (iter.Seq[string], error) {
	fileList := &bytes.Buffer{}
	if err := execGitCmd(
		ctx,
		wd,
		fileList,
		os.Stderr,
		"diff",
		"--name-only",
		"--cached",
	); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(fileList)

	return func(yield func(string) bool) {
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return
			}
		}
	}, nil
}

func Commit(
	ctx context.Context,
	wd,
	msg string,
	args ...string,
) error {
	fmt.Println("MSG", msg)
	fmt.Println("A", len(args), args)

	cmdLine := slices.Concat(
		[]string{"commit"},
		args,
		[]string{fmt.Sprintf("-m\"%s\"", msg)},
	)
	fmt.Println(cmdLine)

	if err := execGitCmd(
		ctx,
		wd,
		os.Stdout,
		os.Stderr,
		cmdLine...,
	); err != nil {
		return err
	}

	return nil
}

func execGitCmd(
	ctx context.Context,
	wd string,
	stdout io.Writer,
	stderr io.Writer,
	args ...string,
) error {
	cmd := exec.CommandContext(
		ctx,
		"git", args...,
	)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = wd

	return cmd.Run()
}
