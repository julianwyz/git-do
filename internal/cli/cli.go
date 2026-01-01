package cli

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/julianwyz/git-do/internal/config"
	"github.com/julianwyz/git-do/internal/credentials"
	"github.com/julianwyz/git-do/internal/llm"
)

type (
	CLI struct {
		Commit Commit `cmd:"" help:"commit"`

		runner   *kong.Context `kong:"-"`
		config   *cliConfig
		userHome string
		cwd      string
	}

	Ctx struct {
		context.Context
		LLM        *llm.LLM
		UserConfig *config.Config
		HomeDir    string
		WorkingDir string
	}
)

func New(opts ...CLIOpt) (*CLI, error) {
	var (
		err      error
		returner = &CLI{
			config: &cliConfig{},
		}
	)

	for _, o := range opts {
		if err = o(returner.config); err != nil {
			return nil, err
		}
	}

	returner.userHome, err = os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	returner.cwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	returner.runner = kong.Parse(
		returner,
		kong.Name("git do"),
		kong.Help(returner.OutputHelp(os.Stdout)),
	)

	return returner, nil
}

func (recv *CLI) Exec(ctx context.Context) error {
	projectConfig, apiCredentials, err := recv.loadConfig()
	if err != nil {
		return err
	}

	llmDriver, err := recv.configureLLM(
		projectConfig, apiCredentials,
	)
	if err != nil {
		return err
	}

	return recv.runner.Run(&Ctx{
		Context:    ctx,
		LLM:        llmDriver,
		HomeDir:    recv.userHome,
		WorkingDir: recv.cwd,
	})
}

func (recv *CLI) FatalIfErrorf(err error, args ...any) {
	recv.runner.FatalIfErrorf(err, args...)
}

func (recv *CLI) configureLLM(
	cfg *config.Config,
	creds *credentials.Credentials,
) (*llm.LLM, error) {
	opts := []llm.LLMOpt{
		llm.WithConfig(cfg),
	}

	if cfg != nil && cfg.LLM != nil {
		if len(cfg.LLM.APIBase) > 0 {
			opts = append(opts, llm.WithAPIBase(cfg.LLM.APIBase))
		}
		if len(cfg.LLM.Model) > 0 {
			opts = append(opts, llm.WithModel(cfg.LLM.Model))
		}
	}

	if creds != nil {
		if len(creds.APIKey) > 0 {
			opts = append(opts, llm.WithAPIKey(creds.APIKey))
		}
	}

	return llm.New(opts...)
}

func (recv *CLI) loadConfig() (
	*config.Config,
	*credentials.Credentials,
	error,
) {
	var (
		err           error
		projectConfig *config.Config
		creds         *credentials.Credentials
	)
	projectConfig, err = config.LoadFrom(
		os.DirFS(recv.cwd),
	)
	if err != nil {
		return nil, nil, err
	}

	if projectConfig.LLM != nil && len(projectConfig.LLM.APIBase) > 0 {
		apiUrl, err := url.Parse(projectConfig.LLM.APIBase)
		// not having a valid url here may bite us later,
		// but for our purposes, we only care about valid urls
		if err == nil {
			creds, err = credentials.LoadFrom(
				os.DirFS(recv.userHome),
				apiUrl.Host,
			)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return projectConfig, creds, nil
}
