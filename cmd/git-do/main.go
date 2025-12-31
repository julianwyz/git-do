package main

import (
	"context"
	"os"
	"time"

	"github.com/julianwyz/git-do/internal/cli"
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

	cli, err := cli.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize cli")
	}

	cli.FatalIfErrorf(
		cli.Exec(ctx),
	)
}
