# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Go course material. Each `lecture-NN/` directory is a self-contained lesson with standalone demo programs (no shared go.mod).

## Build & Run

Programs are standalone scripts — run with `go run main.go` from the lecture subdirectory. There is no project-wide build system, test suite, or linter.

## Language

- Write new code comments in **Russian**.
- Write commit messages in **English**.
- Existing Russian comments — do not rewrite them.

## Structure

Each lecture directory is independent with no shared packages. Do not introduce cross-lecture dependencies.
