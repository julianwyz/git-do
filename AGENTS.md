# Agent Guide

This file is intended to help code agents (e.g. Claude Code, Copilot, Cursor) understand this project and work within it effectively.

## What this project is

`git-do` is a Git CLI extension that adds a `git do` subcommand. It automates common git workflows — primarily commit message generation — using LLMs via any OpenAI-compatible API. The tool is written in Go and intended to be installed globally on a developer's machine.

The project itself uses `git-do` for its own commits (see `.do.toml` and `CONTEXT.md`).

## Commands provided

| Command | Purpose |
|---|---|
| `git do commit` | Generate a commit message for staged changes and commit |
| `git do explain` | Explain changes in a commit or range of commits |
| `git do status` | `git status` with LLM-generated explanations of changes |
| `git do init` | Initialize `git-do` for a project (writes config and credentials) |
| `git do help` | Show help |

## Project layout

```
cmd/git-do/main.go          # Binary entry point — logging setup, CLI init
internal/cli/               # Command implementations
  cli.go                    # Orchestration: arg parsing, config loading, LLM setup
  commit.go                 # `git do commit` logic
  explain.go                # `git do explain` logic
  status.go                 # `git do status` logic
  init.go                   # `git do init` logic (interactive setup)
  help.go                   # Help rendering
  options.go                # Functional options for CLI config
internal/config/            # TOML config loading (.do.toml / do.toml / Dofile)
internal/credentials/       # INI credentials loader (~/.gitdo/credentials)
internal/git/               # Thin wrapper around git CLI commands
internal/llm/               # OpenAI-compatible LLM client + streaming
  prompts/                  # Embedded Go template prompt files (.tmpl.md)
    gen_commit_instruct.tmpl.md
    explain_instruct.tmpl.md
    status_instruct.tmpl.md
```

## Architecture

**Layered, modular design.** The CLI layer orchestrates the other packages; each package has a narrow responsibility.

- **Config** (`internal/config/`): Loads TOML project config from the working directory. Supports multiple file name aliases. Optionally loads a context file (e.g. `CONTEXT.md`) referenced in the config.
- **Credentials** (`internal/credentials/`): Loads an INI file from `~/.gitdo/credentials`. Supports per-domain API keys with a `[default]` fallback.
- **Git** (`internal/git/`): Shells out to `git`. Uses Go 1.22 `iter.Seq2` iterators for lazy diff and commit streaming.
- **LLM** (`internal/llm/`): Uses the official OpenAI Go SDK (`openai-go/v3`). System prompts are embedded Go templates. Streaming is used for `explain` and `status`; non-streaming for `commit` (structured output).
- **CLI** (`internal/cli/`): Uses `kong` for argument parsing. Passes a `Ctx` struct down to each command's `Run` method.

## Key conventions

- **Options pattern**: Both `cli` and `llm` packages expose functional option types for configuration.
- **Error sentinel values**: `ErrNoProjectConfig`, `ErrNoCreds`, `ErrNoPatches` — check for these to produce user-friendly messages.
- **No comments on the obvious**: Comments in this codebase explain *why*, not *what*. Match that style.
- **Iterators**: Diffs and commits are yielded as `iter.Seq2[string, error]` — don't break this pattern when touching `internal/git/`.
- **Embedded prompts**: Prompt templates live in `internal/llm/prompts/` and are embedded via `//go:embed`. Edits there take effect at next build.
- **Version constant**: The canonical version lives in `internal/cli/cli.go` as `Version`. It appears in commit message trailers.

## Configuration files

**Project config** (`.do.toml`, `do.toml`, `Dofile`, or `Dofile.toml` — TOML in all cases):
```toml
version = "1"
language = "en-US"          # BCP 47 tag

[llm]
api_base = "https://api.openai.com/v1"
model = "gpt-4o-mini"

[llm.context]
file = "CONTEXT.md"         # Optional project-specific context for the LLM

[llm.reasoning]
level = "low"               # none | minimal | low | medium | high | xhigh

[commit]
format = "github"           # "github" or "conventional"
```

**Credentials** (`~/.gitdo/credentials` — INI format):
```ini
[default]
api_key = ...

[api.openai.com]
api_key = ...
```

**Environment variables:**
- `GITDO_DEBUG=TRUE` — enables structured debug logging to stderr

## Build and test

```sh
task test      # go test -cover ./...
task lint      # golangci-lint run ./...
go build ./cmd/git-do   # build the binary
```

Go version: **1.25.1** (requires 1.22+ for iterator support).

## Special commit message rules (important for agents)

When the **only** diff is a change to the `Version` constant in `internal/cli/cli.go`, the commit message must be exactly:

```
v<version>
```

No body, no trailers beyond what `git-do` appends automatically.

For all other commits, use the format configured in `.do.toml` (`github` or `conventional`). User-facing CLI changes should be called out explicitly in the commit body; if there are none, omit the callout.

## What agents should not do

- Do not change the credentials file format or location without updating both `internal/credentials/` and `internal/cli/init.go`.
- Do not add direct `fmt.Print` output from library packages (`config`, `credentials`, `git`, `llm`) — output belongs in the `cli` layer.
- Do not skip tests or linting to work around failures — fix the root cause.
- Do not break the `iter.Seq2` interface on git operations; callers rely on lazy evaluation for large diffs.
- Do not hardcode the OpenAI API base — the tool is explicitly LLM-agnostic.
