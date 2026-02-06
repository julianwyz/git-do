package config

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/julianwyz/git-do/internal/git"
	"github.com/julianwyz/git-do/internal/llm"
)

type (
	Config struct {
		Version  string  `toml:"version"`
		Language string  `toml:"language"`
		LLM      *LLM    `toml:"llm"`
		Commit   *Commit `toml:"commit"`

		configFs fs.FS
	}

	LLM struct {
		APIBase   string     `toml:"api_base"`
		Model     string     `toml:"model"`
		Context   *Context   `toml:"context"`
		Reasoning *Reasoning `toml:"reasoning"`
	}

	Reasoning struct {
		Level llm.ReasoningLevel `toml:"level"`
	}

	Commit struct {
		Format git.CommitFormat `toml:"format"`
	}

	Context struct {
		File string `toml:"file"`
	}
)

var (
	ErrNoConfig       = errors.New("no config file found")
	ErrNoContext      = errors.New("no context file available")
	ErrInvalidVersion = errors.New("unknown version specified")

	configFileAliases = [...]string{
		".do.toml",
		"do.toml",
		"Dofile",
		"Dofile.toml",
	}
)

func LoadFrom(fs fs.FS) (*Config, error) {
	for _, variant := range configFileAliases {
		if f, err := fs.Open(variant); err == nil {
			defer f.Close()

			var dst Config
			dec := toml.NewDecoder(f)
			if _, err := dec.Decode(&dst); err != nil {
				return nil, err
			}

			if len(dst.Version) > 0 && dst.Version != "1" {
				// if we have a version, and it is not supported...
				return nil, ErrInvalidVersion
			}

			dst.configFs = fs
			return &dst, nil
		}
	}

	return nil, ErrNoConfig
}

func Exists(f fs.FS) (bool, string) {
	for _, variant := range configFileAliases {
		if _, err := fs.Stat(f, variant); err == nil {
			return true, variant
		}
	}

	return false, ""
}

func WriteDefault(dir string) error {
	def := &Config{
		Version:  "1",
		Language: "en-US",
		LLM: &LLM{
			APIBase: "https://api.openai.com/v1",
			Model:   "gpt-5-mini",
		},
		Commit: &Commit{
			Format: "github",
		},
	}
	content, err := toml.Marshal(def)
	if err != nil {
		return err
	}

	return os.WriteFile(
		filepath.Join(dir, ".do.toml"),
		content,
		0644,
	)
}

func (recv *Config) LoadContextFile() (io.ReadCloser, error) {
	if recv.LLM == nil {
		return nil, ErrNoContext
	}
	if recv.LLM.Context == nil {
		return nil, ErrNoContext
	}
	if len(recv.LLM.Context.File) == 0 {
		return nil, ErrNoContext
	}

	return recv.configFs.Open(recv.LLM.Context.File)
}
