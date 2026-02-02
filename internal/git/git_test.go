package git_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julianwyz/git-do/internal/git"
)

func TestInit(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitdo-test-*")
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := git.Init(t.Context(), dir, &output); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		output.String(),
		"Initialized empty Git repository in",
	) {
		t.Fatal("did not initialize git directory")
	}
}

func TestStatus(t *testing.T) {
	t.Run("unstaged", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		seq, status, err := git.Status(
			t.Context(),
			wd,
			".",
		)
		if err != nil {
			t.Fatal(err)
		}

		if len(status) == 0 {
			t.Fatal("status is empty")
		}

		for f, err := range seq {
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(
				f, "+hello world",
			) {
				t.Fatal("unexpected patch content")
			}

			break
		}
	})

	t.Run("staged", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		seq, status, err := git.Status(
			t.Context(),
			wd,
			".",
		)
		if err != nil {
			t.Fatal(err)
		}

		if len(status) == 0 {
			t.Fatal("status is empty")
		}

		for f, err := range seq {
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(
				f, "+hello world",
			) {
				t.Fatal("unexpected patch content")
			}
		}
	})

	t.Run("renamed", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		if err := os.Rename(
			filepath.Join(wd, "test.txt"),
			filepath.Join(wd, "new-test.txt"),
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"new-test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		seq, status, err := git.Status(
			t.Context(),
			wd,
			".",
		)
		if err != nil {
			t.Fatal(err)
		}

		if len(status) == 0 {
			t.Fatal("status is empty")
		}

		for f, err := range seq {
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(
				f, "+hello world",
			) {

				t.Fatal("unexpected patch content")
			}
		}
	})

	t.Run("commit and change", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world\nfoo bar"),
			0644); err != nil {
			t.Fatal(err)
		}

		seq, status, err := git.Status(
			t.Context(),
			wd,
			".",
		)
		if err != nil {
			t.Fatal(err)
		}

		if len(status) == 0 {
			t.Fatal("status is empty")
		}

		for f, err := range seq {
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(
				f, "+hello world",
			) {

				t.Fatal("unexpected patch content")
			}
		}
	})
}

func TestHeadHash(t *testing.T) {
	wd, err := initNewDir(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"commit",
		"-m",
		"add test",
	); err != nil {
		t.Fatal(err)
	}

	hash, err := git.HeadHash(
		t.Context(),
		wd,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(hash) == 0 {
		t.Fatal("no hash")
	}
}

func TestShowCommit(t *testing.T) {
	wd, err := initNewDir(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"commit",
		"-m",
		"add test",
	); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err = git.ShowCommit(
		t.Context(),
		wd,
		"HEAD",
		&output,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		output.String(),
		"add test",
	) {
		t.Fatal("no commit info")
	}
}

func TestDiffsOfCommit(t *testing.T) {
	t.Run("root parent", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		hash, err := git.HeadHash(
			t.Context(),
			wd,
		)
		if err != nil {
			t.Fatal(err)
		}

		var output bytes.Buffer
		err = git.DiffsOfCommit(
			t.Context(),
			wd,
			hash,
			"test.txt",
			&output,
		)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(
			output.String(),
			"+hello world",
		) {
			t.Fatal("no commit info")
		}
	})

	t.Run("history", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world\nfoo bar"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		hash, err := git.HeadHash(
			t.Context(),
			wd,
		)
		if err != nil {
			t.Fatal(err)
		}

		var output bytes.Buffer
		err = git.DiffsOfCommit(
			t.Context(),
			wd,
			hash,
			"test.txt",
			&output,
		)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(
			output.String(),
			"+foo bar",
		) {
			t.Fatal("no commit info")
		}
	})
}

func TestStagedDiffs(t *testing.T) {
	wd, err := initNewDir(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := git.StagedDiffs(
		t.Context(),
		wd,
		"test.txt",
		&output,
	); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		output.String(),
		"+hello world",
	) {
		t.Fatal("no commit info")
	}
}

func TestListStaged(t *testing.T) {
	wd, err := initNewDir(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	seq, err := git.ListStaged(
		t.Context(),
		wd,
	)
	if err != nil {
		t.Fatal(err)
	}

	var items []string
	for item, err := range seq {
		if err != nil {
			t.Fatal(err)
		}

		items = append(items, item)
	}

	if len(items) != 1 {
		t.Fatal("bad length")
	}

	if !strings.Contains(
		items[0],
		"+hello world",
	) {
		t.Fatal("no commit info")
	}
}

func TestCommit(t *testing.T) {
	wd, err := initNewDir(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	if err := git.Commit(
		t.Context(),
		wd,
		strings.NewReader("hello world"),
		"--no-verify",
	); err != nil {
		t.Fatal(err)
	}
}

func TestListCommitChanges(t *testing.T) {
	wd, err := initNewDir(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"commit",
		"-m",
		"add test",
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(wd, "test.txt"),
		[]byte("hello world\nfoo bar"),
		0644); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"add",
		"test.txt",
	); err != nil {
		t.Fatal(err)
	}

	if err := runGitCmd(
		t.Context(),
		wd,
		"commit",
		"-m",
		"add test",
	); err != nil {
		t.Fatal(err)
	}

	hash, err := git.HeadHash(
		t.Context(),
		wd,
	)
	if err != nil {
		t.Fatal(err)
	}

	seq, err := git.ListCommitChanges(
		t.Context(),
		wd,
		hash,
	)
	if err != nil {
		t.Fatal(err)
	}

	var items []string
	for item, err := range seq {
		if err != nil {
			t.Fatal(err)
		}

		items = append(items, item)
	}

	if len(items) != 1 {
		t.Fatal("bad length")
	}

	if !strings.Contains(
		items[0],
		"+hello world",
	) {
		t.Fatal("no commit info")
	}
}

func TestCommitsBetween(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world\nfoo bar"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		hash, err := git.HeadHash(
			t.Context(),
			wd,
		)
		if err != nil {
			t.Fatal(err)
		}

		seq, err := git.CommitsBetween(
			t.Context(),
			wd,
			[]string{hash},
		)
		if err != nil {
			t.Fatal(err)
		}

		var items []string
		for item, err := range seq {
			if err != nil {
				t.Fatal(err)
			}

			items = append(items, item)
		}

		if len(items) != 1 {
			t.Fatal("bad length")
		}

		if !strings.Contains(
			items[0],
			"+hello world",
		) {
			t.Fatal("no commit info")
		}
	})

	t.Run("multi", func(t *testing.T) {
		wd, err := initNewDir(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(wd, "test.txt"),
			[]byte("hello world\nfoo bar"),
			0644); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"add",
			"test.txt",
		); err != nil {
			t.Fatal(err)
		}

		if err := runGitCmd(
			t.Context(),
			wd,
			"commit",
			"-m",
			"add test",
		); err != nil {
			t.Fatal(err)
		}

		hash, err := git.HeadHash(
			t.Context(),
			wd,
		)
		if err != nil {
			t.Fatal(err)
		}

		r, _ := git.RootCommit(t.Context(), wd)
		seq, err := git.CommitsBetween(
			t.Context(),
			wd,
			[]string{r, hash},
		)
		if err != nil {
			t.Fatal(err)
		}

		var items []string
		for item, err := range seq {
			if err != nil {
				t.Fatal(err)
			}

			items = append(items, item)
		}

		if len(items) != 2 {
			t.Fatal("bad length")
		}

		if !strings.Contains(
			items[0],
			"+hello world",
		) {
			t.Fatal("no commit info")
		}
	})
}

func initNewDir(ctx context.Context) (string, error) {
	dir, err := os.MkdirTemp("", "gitdo-test-*")
	if err != nil {
		return "", err
	}

	if err := git.Init(ctx, dir, nil); err != nil {
		return "", err
	}

	return dir, nil
}

func runGitCmd(
	ctx context.Context,
	wd string,
	args ...string,
) error {
	cmd := exec.CommandContext(
		ctx,
		"git", args...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = wd

	return cmd.Run()
}
