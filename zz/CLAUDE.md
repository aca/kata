# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go CLI tool called `zz` that wraps commands to run them in the background using systemd-run. The tool passes the current environment variables to the systemd service since systemd-run doesn't inherit the environment by default.

## Build and Run

- Build: `go build -o zz`
- Run: `./zz <command> [args...]`
- List running commands: `./zz --list`
- Example: `./zz sleep 10` (runs `sleep 10` in background via systemd)

## Features

- Runs commands in background using systemd-run
- Shows desktop notifications when commands complete
- Shows critical notification when commands fail
- Lists running commands with `--list` flag

## Architecture

The tool uses `systemd-run --user --scope` to create transient systemd scopes for running commands. All environment variables from the current shell are passed using `--setenv` flags to ensure the background process has access to the same environment.

## Key Implementation Details

- Uses `os.Environ()` to capture all environment variables
- Passes each env var with `--setenv` flag to systemd-run
- Preserves exit codes from the executed command
- Forwards stdin/stdout/stderr to maintain interactive capabilities