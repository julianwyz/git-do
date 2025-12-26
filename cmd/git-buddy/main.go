package main

import (
	"context"

	"github.com/julianwyz/git-buddy/internal/cli"
)

func main() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	cli := cli.New()
	err := cli.Exec(ctx)

	cli.FatalIfErrorf(err)
}
