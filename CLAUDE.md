# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Go library for rendering annotated source code with line numbers and labeled markers. Single-package library (`package annotate`).

## Toolchain

Managed via mise. Run tools through `mise exec --` if not on PATH.

- Go 1.26.1
- golangci-lint 2.11.3

## Commands

```bash
# Run all tests
go test -v

# Run a single test
go test -run TestRender_SpecExample -v

# Lint
golangci-lint run
```

## Architecture

All code lives in `annotate.go` (single file, single package). The rendering pipeline:

1. **Parse lines** — split source bytes by `\n`, record byte offsets, expand tabs (fixed width 4)
2. **Map labels to lines** — for each label, determine which lines its Span covers; multi-line spans are split into per-line sub-spans
3. **Generate output** — for each line, emit `{line number} | {text}` followed by marker lines for any labels on that line

Key types:
- `Span` — byte offset range `[Start, End)` into source
- `Label` — a Span + marker character + annotation text
- `Renderer` — entry point; `Render()` returns string, `Write()` writes to `io.Writer`

Display width (CJK wide characters) is handled via `go-runewidth`. Marker positioning uses display columns, not byte counts.
