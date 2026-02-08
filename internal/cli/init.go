package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/julianwyz/git-do/internal/config"
	"github.com/julianwyz/git-do/internal/credentials"
	"github.com/julianwyz/git-do/internal/git"
	"golang.org/x/term"
)

type (
	Init struct{}
)

const (
	initHelp = `git do init
=======

This will create the user credentials file in your home directory, it it has not yet been created.

It will also create a ` + "`git do`" + ` configuration file in your current working directory.
If a ` + "`git`" + ` repo has not yet been created in the directory, ` + "`git init`" + ` will be called as well.

Flags:

` + "`-h`" + `, ` + "`--help`" + `
> Show this help message.
`
)

func (recv *Init) Run(ctx *Ctx) error {
	if !recv.hasGit(ctx.WorkingDir) {
		if err := git.Init(ctx, ctx.WorkingDir, ctx.Output); err != nil {
			return err
		}
		_, _ = ctx.Output.WriteString("Established git repo.\n")
	} else {
		_, _ = ctx.Output.WriteString("git repo already established.\n")
	}

	if e, f := config.Exists(
		os.DirFS(ctx.WorkingDir),
	); !e {
		if err := config.WriteDefault(ctx.WorkingDir); err != nil {
			return err
		}
		_, _ = ctx.Output.WriteString("Created initial project configuration.\n")
	} else {
		_, _ = fmt.Fprintf(ctx.Output, "%s configuration file already exists.\n", f)
	}

	if e := credentials.Exists(
		os.DirFS(ctx.HomeDir),
	); !e {
		_, _ = ctx.Output.WriteString("No credentials file found. Let's make a new one!.\n")

		var key string
		if term.IsTerminal(int(os.Stdout.Fd())) {
			if err :=
				huh.NewInput().
					Title("What should we use for the default LLM API Key?").
					Description("If you don't need an API Key, you can just leave it blank.").
					Value(&key).
					Run(); err != nil {
				return err
			}
		} else {
			_, _ = ctx.Output.WriteString("Device is non-interactive. Using empty key in credentials. You will have to update this manually.\n")
		}

		credFile, err := credentials.WriteDefault(
			ctx.HomeDir,
			key,
		)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(ctx.Output, "Wrote credentials file to: %s\n", credFile)
	} else {
		_, _ = ctx.Output.WriteString("Credentials file already exists.\n")
	}

	_, _ = ctx.Output.WriteString("\n🎉 Initial setup complete. Happy coding!\n")

	return nil
}

func (recv Init) Help(dst io.Writer) error {
	return renderHelpMarkdown(dst, initHelp)
}

func (recv *Init) hasGit(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))

	return err == nil
}
