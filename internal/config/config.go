package config

import (
	"errors"
	"io"
	"io/fs"

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
		"do.toml",
		"Dofile",
		"Dofile.toml",
		".do.toml",
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
