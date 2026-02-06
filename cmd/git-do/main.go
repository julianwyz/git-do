package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/julianwyz/git-do/internal/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	var dst = io.Discard

	if v, set := os.LookupEnv("GITDO_DEBUG"); set && v == "TRUE" {
		dst = os.Stderr
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        dst,
		TimeFormat: time.RFC3339,
	})
}

func main() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	runner, err := cli.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize cli")
	}

	exitCode := 0
	if err := runner.Exec(ctx); err != nil {
		exitCode = 1

		switch {
		case errors.Is(err, cli.ErrNoCreds):
			_, _ = os.Stdout.WriteString("No user credentials found. Have you ran `git do init` yet?\n")
		case errors.Is(err, cli.ErrNoProjectConfig):
			_, _ = os.Stderr.WriteString("No project configuration file found in current directory. Have you ran `git do init` yet?\n")
		default:
			fmt.Fprintf(os.Stderr, "Encountered unknown error: %s\n", err.Error())
		}
	}

	os.Exit(exitCode)
}
