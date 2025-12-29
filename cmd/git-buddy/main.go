package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/julianwyz/git-buddy/internal/cli"
	"github.com/julianwyz/git-buddy/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}

	userConfig, err := config.Load(
		filepath.Join(home, "buddy.toml"),
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	cli, err := cli.New(
		cli.WithConfig(userConfig),
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	cli.FatalIfErrorf(
		cli.Exec(ctx),
	)
}
