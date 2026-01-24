package credentials

import (
	"errors"
	"io/fs"
	"os"
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

func WriteDefault(dir, key string) (string, error) {
	cfg := ini.Empty()
	s, err := cfg.NewSection("default")
	if err != nil {
		return "", err
	}
	if _, err := s.NewKey(apiKeyFieldName, key); err != nil {
		return "", err
	}

	_ = os.Mkdir(
		filepath.Join(dir,
			".gitdo",
		),
		os.ModeDir,
	)
	w, err := os.Create(
		filepath.Join(dir,
			".gitdo",
			"credentials",
		),
	)
	if err != nil {
		return "", err
	}
	defer w.Close()

	_, err = cfg.WriteTo(w)
	return w.Name(), err
}

func Exists(f fs.FS) bool {
	_, err := fs.Stat(
		f,
		filepath.Join(".gitdo", "credentials"),
	)

	return err == nil
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
