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

Single package (`package annotate`), three files:

- `annotate.go` — rendering pipeline (`Renderer`, `Render`, `Write`)
- `style.go` — style type definitions (`StyleFunc`, `Style`, `LabelStyle`), ANSI StyleFunc variables, presets, `ComposeStyles`
- `option.go` — functional options (`Option`, `WithStyle`, `WithBefore`, `WithAfter`)

### Rendering pipeline

1. **Parse lines** — split source bytes by `\n`, record byte offsets, expand tabs (fixed width 4)
2. **Map labels to lines** — for each label, determine which lines its Span covers; multi-line spans are split into per-line sub-spans
3. **Compute visible lines** — only lines covered by labels (plus Before/After context) are included; if no labels, output is empty
4. **Generate output** — for each visible line, emit `{line number} | {text}` followed by marker lines for any labels on that line. Non-contiguous line groups are separated by a styled `...` ellipsis. Style is applied after all width/position calculations. Line number width is based on the maximum visible line number.

### Key types

- `Span` — byte offset range `[Start, End)` into source
- `Label` — a Span + marker character + annotation text + optional `LabelStyle` + optional `Before`/`After` (`*int`) context override
- `Renderer` — entry point; `Render()` returns string, `Write()` writes to `io.Writer`. Has `Before`/`After` fields for default context lines.
- `StyleFunc` — `func(string) string` callback for styling output elements
- `Style` — global style config (7 elements: LineNumber, Separator, SpanCode, NonSpanCode, Marker, LabelText, Ellipsis)
- `LabelStyle` — per-label style override (5 elements, no NonSpanCode or Ellipsis)

Display width (CJK wide characters) is handled via `go-runewidth`. Marker positioning uses display columns, not byte counts.
