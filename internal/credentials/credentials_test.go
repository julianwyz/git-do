package credentials_test

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/julianwyz/git-do/internal/credentials"
)

var (
	//go:embed all:fixtures
	fixtures embed.FS
)

func TestLoadFrom(t *testing.T) {
	t.Run("blank", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "blank"),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = credentials.LoadFrom(sub, "example.com")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, credentials.ErrDomain) {
			t.Fatal("incorrect error type")
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

		_, err = credentials.LoadFrom(sub, "example.com")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatal("incorrect error type")
		}
	})

	t.Run("bad ini", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "syntax"),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = credentials.LoadFrom(sub, "example.com")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("default", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "ok"),
		)
		if err != nil {
			t.Fatal(err)
		}

		creds, err := credentials.LoadFrom(sub, "not-known-domain.foo")
		if err != nil {
			t.Fatal("unexpected error")
		}

		if creds.APIKey != "default key" {
			t.Fatal("unexpected credential value")
		}
	})

	t.Run("localhost", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "ok"),
		)
		if err != nil {
			t.Fatal(err)
		}

		creds, err := credentials.LoadFrom(sub, "localhost")
		if err != nil {
			t.Fatal("unexpected error")
		}

		if creds.APIKey != "localhost key" {
			t.Fatal("unexpected credential value")
		}
	})

	t.Run("localhost with port", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "ok"),
		)
		if err != nil {
			t.Fatal(err)
		}

		creds, err := credentials.LoadFrom(sub, "localhost:11434")
		if err != nil {
			t.Fatal("unexpected error")
		}

		if creds.APIKey != "localhost port key" {
			t.Fatal("unexpected credential value")
		}
	})

	t.Run("domain", func(t *testing.T) {
		sub, err := fs.Sub(
			fixtures,
			filepath.Join("fixtures", "ok"),
		)
		if err != nil {
			t.Fatal(err)
		}

		creds, err := credentials.LoadFrom(sub, "api.example.com")
		if err != nil {
			t.Fatal("unexpected error")
		}

		if creds.APIKey != "example key" {
			t.Fatal("unexpected credential value")
		}
	})
}

func TestExists(t *testing.T) {
	sub, err := fs.Sub(
		fixtures,
		filepath.Join("fixtures", "ok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	sub2, err := fs.Sub(
		fixtures,
		filepath.Join("fixtures", "empty"),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !credentials.Exists(sub) {
		t.Fatal("should exist")
	}

	if credentials.Exists(sub2) {
		t.Fatal("should not exist")
	}
}

func TestWriteDefault(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitdo-test-*")
	if err != nil {
		t.Fatal(err)
	}

	fileLoc, err := credentials.WriteDefault(
		dir, "my_super_secret_key",
	)
	if err != nil {
		t.Fatal(err)
	}

	if fileLoc != filepath.Join(
		dir, ".gitdo", "credentials",
	) {
		t.Fatal("wrong path")
	}

	// try reading it

	creds, err := credentials.LoadFrom(
		os.DirFS(dir),
		"example.com",
	)
	if err != nil {
		t.Fatal(err)
	}

	if creds.APIKey != "my_super_secret_key" {
		t.Fatal("wrong credential value")
	}
}
