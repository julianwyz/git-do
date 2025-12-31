package credentials

import (
	"errors"
	"io/fs"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type (
	Credentials struct {
		APIKey string
	}
)

var (
	ErrDomain = errors.New("no credentials available for the domain")
)

const (
	apiKeyFieldName = "api_key"
)

func LoadFrom(fs fs.FS, domain string) (*Credentials, error) {
	f, err := fs.Open(
		filepath.Join(".gitdo", "credentials"),
	)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, err := ini.Load(f)
	if err != nil {
		return nil, err
	}

	returner := &Credentials{}
	if err := retrieveApiKey(
		cfg, domain, returner,
	); err != nil {

	}

	return returner, nil
}

func retrieveApiKey(cfg *ini.File, domain string, dst *Credentials) error {
	var apiKey *ini.Key
	useSection := func(s string) {
		if s, err := cfg.GetSection(s); err == nil {
			apiKey, _ = s.GetKey(apiKeyFieldName)
		}
	}

	useSection(domain)
	if apiKey == nil {
		useSection("default")
	}

	if apiKey == nil {
		// no default and no domain keys present in config
		return ErrDomain
	}

	dst.APIKey = apiKey.String()

	return nil
}
