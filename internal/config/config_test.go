package config_test

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/julianwyz/git-do/internal/config"
)

var (
	//go:embed all:fixtures
	fixtures embed.FS
)

func TestLoadFrom(t *testing.T) {
	variants := []string{
		"1",
		"2",
		"3",
		"4",
		"prio",
	}

	for _, v := range variants {
		t.Run(v, func(t *testing.T) {
			sub, err := fs.Sub(
				fixtures,
				filepath.Join("fixtures", v),
			)
			if err != nil {
				t.Fatal(err)
			}

			c, err := config.LoadFrom(sub)
			if err != nil {
				t.Fatal(err)
			}

			if c.Version != "1" {
				t.Fatal("bad version")
			}

			if c.Language != "en-US" {
				t.Fatal("bad lang")
			}
		})
	}

	t.Run("bad version", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "bad_version"),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = config.LoadFrom(sub)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, config.ErrInvalidVersion) {
			t.Fatal("incorrect error")
		}
	})

	t.Run("empty", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "empty"),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = config.LoadFrom(sub)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, config.ErrNoConfig) {
			t.Fatal("incorrect error")
		}
	})

	t.Run("bad toml", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "syntax"),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = config.LoadFrom(sub)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestExists(t *testing.T) {
	variants := [][]string{
		{"1", ".do.toml"},
		{"2", "do.toml"},
		{"3", "Dofile"},
		{"4", "Dofile.toml"},
	}

	for _, v := range variants {
		t.Run(v[0], func(t *testing.T) {
			sub, err := fs.Sub(
				fixtures,
				filepath.Join("fixtures", v[0]),
			)
			if err != nil {
				t.Fatal(err)
			}

			exists, path := config.Exists(sub)
			if !exists {
				t.Fatal("expected config to exist")
			}

			if path != v[1] {
				t.Fatalf("expected path to equal %s", v[1])
			}
		})
	}

	t.Run("empty", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "empty"),
		)
		if err != nil {
			t.Fatal(err)
		}

		exists, _ := config.Exists(sub)
		if exists {
			t.Fatal("expected not to find config")
		}
	})
}

func TestWriteDefault(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitdo-test-*")
	if err != nil {
		t.Fatal(err)
	}

	err = config.WriteDefault(
		dir,
	)
	if err != nil {
		t.Fatal(err)
	}

	exists, path := config.Exists(
		os.DirFS(dir),
	)
	if !exists {
		t.Fatal("expected to find config")
	}

	if path != ".do.toml" {
		t.Fatal("expected .do.toml file")
	}
}
