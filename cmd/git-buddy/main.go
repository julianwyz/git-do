package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/julianwyz/git-buddy/internal/cli"
	"github.com/julianwyz/git-buddy/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func main() {

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to determine home directory")
	}

	userConfig, err := config.Load(
		filepath.Join(home, "buddy.toml"),
	)
	if err != nil {
		log.Fatal().Err(err).
			Msg("cannot find configuration file in your home directory")
	}

	cli, err := cli.New(
		cli.WithConfig(userConfig),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize cli")
	}

	cli.FatalIfErrorf(
		cli.Exec(ctx),
	)
}
