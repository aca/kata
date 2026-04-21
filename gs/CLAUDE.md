# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`gs` is a single-file Go CLI that clones a git repo into a path derived from its URL and default branch:

    gs github.com/xtdlib/pgmap  →  ~/src/github.com/xtdlib/pgmap/<default-branch>

The default branch is detected at runtime with `git ls-remote --symref <url> HEAD` (no GitHub API, works for any git host). The clone itself shells out to `git clone --branch <branch>`.

Input forms accepted by `main.go`: bare `host/owner/repo`, `https://...`, and `git@host:owner/repo`. All are normalized to `https://host/owner/repo` before use; a trailing `.git` is stripped.

## Build & run

    go build -o gs .
    ./gs <repo>

There are no tests, no linter config, and no dependencies beyond the standard library — do not add any without a concrete reason.
