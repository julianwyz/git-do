# Overview

This project is called `git-do`. It is an addon for the `git` CLI that adds a `git do` subcommand.

`git-do` provides an easy way to automate common git commands like `commit` by leveraging generative AI.

# Commit message generation

When generating commit messages using `git do commit`, be sure to analyze the diffs and make a distinction between changes to this project's internals and any user-facing changes.

Be sure to explicitly call out user-facing CLI changes for consumer's of this project to be aware of.

# Use-case

A user downloads and installs the `git-do` addon to their machine then can use the `git do` command in their own git projects.

The `git-do` CLI allows for configuration options to be set using a config file at the root of a git project.

This file may be named one of the following:

- `do.toml`
- `Dofile`
- `Dofile.toml`
- `.do.toml`

Regardless of name or extension, the configuration language support is TOML.

In addition to this project being used by other git repositories, this project _itself_ contains a `do.toml` file that is used to use `git-do` in its own git repository.

# Project technical specifications

|  |  |
|---|---|
| Programming Language | Golang |
| Development task runner | `task` / Taskfile |