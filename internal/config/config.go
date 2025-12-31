package config

import (
	"errors"
	"io/fs"

	"github.com/BurntSushi/toml"
)

type (
	Config struct {
		Version  string  `toml:"version"`
		Language string  `toml:"language"`
		LLM      *LLM    `toml:"llm"`
		Commit   *Commit `toml:"commit"`
	}

	LLM struct {
		APIBase string `toml:"api_base"`
		Model   string `toml:"model"`
	}

	Commit struct {
		Format CommitFormat `json:"format"`
	}

	CommitFormat string
)

const (
	GithubCommitFormat       = CommitFormat("github")
	ConventionalCommitFormat = CommitFormat("conventional")
)

var (
	ErrNoConfig       = errors.New("no config file found")
	ErrInvalidVersion = errors.New("unknown version specified")

	configFileAliases = [...]string{
		"do.toml",
		"Dofile",
		"Dofile.toml",
		".do.toml",
	}
)

func Load(fp string) (*Config, error) {
	var dst Config
	if _, err := toml.DecodeFile(fp, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

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

			return &dst, nil
		}
	}

	return nil, ErrNoConfig
}
