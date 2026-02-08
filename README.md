# git-do

`git-do` is an extension for `git` to give your commit messages some extra ✨ pizzazz ✨ with the help of AI.

> [!NOTE]  
> `git do` is still experimental and subject to change. While I don't expect there to be any _huge_ breaking changes, be aware that there _may_ be breaking changes introduced until a formal `v1` is released.

## What does (g)it _do_?

`git-do` adds a `do` subcommand to your `git` CLI. It connects to any Large Language Model API that conforms to the [OpenAI API Spec](https://github.com/openai/openai-openapi) to automate and enhance common git workflows.

### Commands

| Command          | What does it do?                                                                   |
| ---------------- | ---------------------------------------------------------------------------------- |
| `git do commit`  | Generate a commit message of your staged changes and automatically commit.         |
| `git do explain` | Explain the changes made in a commit, or range of commits.                         |
| `git do init`    | Initialize the `git do` tool and setup the project config file.                    |
| `git do status`  | Enhanced version of `git status` that includes a brief explanation of the changes. |

You can see all, detailed, usage information by running `git do help`.

## Installing

### From source

Compile the CLI using Go:

```sh
go install github.com/julianwyz/git-do/cmd/git-do@latest
```

Then move the binary to your system's `$PATH`. It will then be available at `git do ...`.

## Getting started

Once you've installed the CLI, run `git do init` in your project's directory. This will setup the project-level configuration, your [user credentials](#credentials-file) for accessing LLM services and will even initialize an empty `git` repo if you haven't done so already.

### Configuration

After running `git do init`, a `.do.toml` file will be created in your directory.

The available options are:

```toml
# The config file version. "1" is the only accepted value.
version = "1"

# A BCP 47 language tag that will be provided to the LLM.
# When used with multi-lingual models, all generated content will be in this language.
language = "en-US"

[llm]
# The base URL to access the LLM API.
api_base = "https://api.openai.com/v1"
# The model to use.
model = "gpt-5-mini"

[llm.context]
# An optional file that will be provided to the LLM to provide
# context on your project and to tune responses.
file = "CONTEXT.md"

[llm.reasoning]
# Optionally specify the intensity of reasoning models.
level = "low"

[commit]
# The commit message standard to use.
# Supported values: "github", "conventional"
format = "github"
```

#### LLM Configuration

`git do` utilizes the OpenAI API standard. Any API that conforms to this standard may be used, including local models through tools like [Ollama](https://ollama.com/).

### Credentials file

The `git do` credentials file is located at: `$HOME/.gitdo/credentials`.

This file is interpreted as an [INI file](https://en.wikipedia.org/wiki/INI_file). For example:

```ini
[default]
api_key = hello_world
```

The `default` section will be used for all API calls to the LLM API unless there is a matching hostname defined in your credentials.

When interacting with a `git do` project that is configured with an `api_base` of `https://api.openai.com/v1`, a section with the key `api.openai.com` will be used instead.

```ini
[default]
api_key = hello_world

[api.openai.com]
api_key = something_else
```

This allows you to easily use `git do` in multiple projects with multiple LLM providers simultaneously.

## Motivation

> [Commit messages to me are almost as important as the code change itself.](<(https://linux.slashdot.org/story/20/07/03/2133201/linus-torvalds-i-do-no-coding-any-more#:~:text=commit%20messages%20to%20me%20are%20almost%20as%20important%20as%20the%20code%20change%20itself.)>)
>
> \- Linus Torvalds

We all know that git commit messages are important. A good commit message can help document your code, educate new-comers to a codebase and is an all-around _good thing_.

We also all know that commit messages have a tendency to get... well... I think [XKCD summed it up best](https://xkcd.com/1296/):

![xkcd-1296](./.github/assets/xkcd-1296.png)

This is the core motivation behind `git-do`. It is designed to easily integrate into your workflow and help you craft more robust, accurate and informative commit messages.

## Contributing

Thanks for considering contributing to this project! Feel free to open Github Issues and/or Pull Requests with any contribution you may have. More info is available in [CONTRIBUTING.md](./CONTRIBUTING.md).
