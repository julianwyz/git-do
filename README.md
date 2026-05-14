# git-do

`git-do` is a Git CLI extension that automates commit messages and explains code changes using LLMs. It adds a `do` subcommand to `git` and works with any OpenAI-compatible API, including OpenAI, Anthropic, and local models via Ollama.

> [!NOTE]
> `git do` is still experimental and subject to change. While I don't expect there to be any _huge_ breaking changes, be aware that there _may_ be breaking changes introduced until a formal `v1` is released.

## What does (g)it _do_?

`git-do` connects to an LLM to handle the parts of your git workflow that benefit from understanding code: writing commit messages, explaining diffs, and summarizing what's changed.

### Commands

| Command          | What it does                                                                   |
| ---------------- | ------------------------------------------------------------------------------ |
| `git do commit`  | Generate a commit message for staged changes and commit.                       |
| `git do explain` | Explain the changes in a commit or range of commits.                           |
| `git do status`  | `git status` with an LLM-generated summary of what changed and why it matters. |
| `git do init`    | Set up `git-do` for a project (creates the config file and credentials entry). |
| `git do help`    | Show detailed usage for all commands.                                          |

## Installing

```sh
go install github.com/julianwyz/git-do/cmd/git-do@latest
```

The binary name is `git-do`. Because Git treats any `git-<subcommand>` binary on your `$PATH` as a subcommand, it becomes available as `git do` automatically.

## Getting started

### 1. Initialize a project

Run `git do init` inside any Git repository. It walks you through picking an LLM provider and model, writes a `.do.toml` config file to your project, and sets up your API key in `~/.gitdo/credentials`.

```sh
cd my-project
git do init
```

### 2. Stage your changes

`git-do` works on whatever is staged, just like `git commit`.

```sh
git add -p   # or git add <files>
```

### 3. Generate a commit

```sh
git do commit
```

`git-do` sends your staged diff and optional project context to the LLM, receives a structured commit message, prints it for review, and commits. The message follows whichever format you configured (`github` or `conventional`).

### Explaining commits

```sh
# Explain the last commit
git do explain HEAD

# Explain a range of commits
git do explain main..my-branch

# Explain a specific commit by hash
git do explain abc1234
```

### Enhanced status

```sh
git do status
```

Runs `git status` and appends an LLM-generated plain-English summary of what's changed. Useful for picking up a half-finished branch or reviewing before a PR.

## Configuration

### Project config (`.do.toml`)

After `git do init`, a `.do.toml` appears in your project directory. Commit this file alongside your code; it defines how `git-do` behaves for everyone working in the repo.

```toml
# Config file version. "1" is the only accepted value.
version = "1"

# BCP 47 language tag. All LLM output will be in this language.
language = "en-US"

[llm]
# Base URL for the LLM API.
api_base = "https://api.openai.com/v1"
# Model to use.
model = "gpt-4o-mini"

[llm.context]
# Optional: a file whose contents are sent to the LLM as project context.
# Use this to tune responses: describe your team's conventions, the codebase
# domain, commit message preferences, etc.
file = "CONTEXT.md"

[llm.reasoning]
# Reasoning effort for models that support it.
# Values: none | minimal | low | medium | high | xhigh
level = "low"

[commit]
# Commit message format.
# "github"       - subject line + bullet-point body
# "conventional" - Conventional Commits (feat:, fix:, chore:, etc.)
format = "github"
```

#### Project context file

The `[llm.context]` file is one of the most useful features. Point it at any Markdown file and its contents are included in every prompt. Use it to tell the LLM things like:

- What your project does and its domain terminology
- Your team's commit message conventions or preferences
- Common patterns to follow or avoid

The project itself uses an `AGENTS.md` file for this; see [AGENTS.md](./AGENTS.md) for a real example.

#### Supported LLM providers

`git-do` works with any API that conforms to the OpenAI spec, plus Anthropic natively:

| Provider                  | `api_base`                  |
| ------------------------- | --------------------------- |
| OpenAI                    | `https://api.openai.com/v1` |
| Anthropic                 | `https://api.anthropic.com` |
| Ollama (local)            | `http://localhost:11434/v1` |
| Any OpenAI-compatible API | your endpoint               |

### Credentials (`~/.gitdo/credentials`)

API keys live in `~/.gitdo/credentials` as an INI file, separate from your project config so keys are never committed to source control.

```ini
[default]
api_key = sk-...
```

You can store keys for multiple providers. The section key is the hostname of the `api_base`:

```ini
[default]
api_key = sk-...

[api.openai.com]
api_key = sk-openai-...

[api.anthropic.com]
api_key = sk-ant-...
```

When `git-do` makes a request, it looks for a credentials section matching the `api_base` hostname. If none matches, it falls back to `[default]`. This lets you use different providers across different projects without any extra setup.

## Motivation

> [Commit messages to me are almost as important as the code change itself.](https://linux.slashdot.org/story/20/07/03/2133201/linus-torvalds-i-do-no-coding-any-more#:~:text=commit%20messages%20to%20me%20are%20almost%20as%20important%20as%20the%20code%20change%20itself.)
>
> \- Linus Torvalds

We all know that git commit messages are important. A good commit message documents intent, educates newcomers, and makes `git log` useful months later.

We also all know how they tend to go in practice. [XKCD summed it up best](https://xkcd.com/1296/):

![xkcd-1296](./.github/assets/xkcd-1296.png)

`git-do` is designed to fit naturally into an existing workflow: no new tooling layer, no UI, just an extra subcommand. Stage your changes, run `git do commit`, and get a message that actually says something.

## Contributing

Thanks for considering contributing! Feel free to open GitHub Issues and Pull Requests. More info is in [CONTRIBUTING.md](./CONTRIBUTING.md).
