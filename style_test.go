package annotate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnsiStyleFuncs(t *testing.T) {
	tests := []struct {
		name     string
		fn       StyleFunc
		input    string
		expected string
	}{
		{"Bold", Bold, "hello", "\033[1mhello\033[0m"},
		{"Dim", Dim, "hello", "\033[2mhello\033[0m"},
		{"Italic", Italic, "hello", "\033[3mhello\033[0m"},
		{"Underline", Underline, "hello", "\033[4mhello\033[0m"},
		{"FgRed", FgRed, "hello", "\033[31mhello\033[0m"},
		{"FgGreen", FgGreen, "hello", "\033[32mhello\033[0m"},
		{"FgYellow", FgYellow, "hello", "\033[33mhello\033[0m"},
		{"FgBlue", FgBlue, "hello", "\033[34mhello\033[0m"},
		{"FgMagenta", FgMagenta, "hello", "\033[35mhello\033[0m"},
		{"FgCyan", FgCyan, "hello", "\033[36mhello\033[0m"},
		{"FgWhite", FgWhite, "hello", "\033[37mhello\033[0m"},
		{"BgRed", BgRed, "hello", "\033[41mhello\033[0m"},
		{"BgGreen", BgGreen, "hello", "\033[42mhello\033[0m"},
		{"BgYellow", BgYellow, "hello", "\033[43mhello\033[0m"},
		{"BgBlue", BgBlue, "hello", "\033[44mhello\033[0m"},
		{"BgMagenta", BgMagenta, "hello", "\033[45mhello\033[0m"},
		{"BgCyan", BgCyan, "hello", "\033[46mhello\033[0m"},
		{"BgWhite", BgWhite, "hello", "\033[47mhello\033[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.fn(tt.input))
		})
	}
}

func TestComposeStyles(t *testing.T) {
	t.Run("composes multiple styles", func(t *testing.T) {
		fn := ComposeStyles(FgRed, Bold)
		got := fn("hello")
		assert.Equal(t, "\033[31m\033[1mhello\033[0m\033[0m", got)
	})

	t.Run("single style", func(t *testing.T) {
		fn := ComposeStyles(Bold)
		got := fn("hello")
		assert.Equal(t, "\033[1mhello\033[0m", got)
	})

	t.Run("no styles returns passthrough", func(t *testing.T) {
		fn := ComposeStyles()
		got := fn("hello")
		assert.Equal(t, "hello", got)
	})

	t.Run("three styles", func(t *testing.T) {
		fn := ComposeStyles(FgRed, Bold, Underline)
		got := fn("hello")
		assert.Equal(t, "\033[31m\033[1m\033[4mhello\033[0m\033[0m\033[0m", got)
	})
}
