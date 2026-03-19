package annotate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender_EmptySource(t *testing.T) {
	r := New()
	got, err := r.Render([]byte{}, nil)
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestRender_SingleLineNoLabels(t *testing.T) {
	r := New()
	got, err := r.Render([]byte("foo: bar"), nil)
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestRender_SingleLineWithLabel(t *testing.T) {
	r := New()
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 0, End: 8}, Marker: MarkerDash, Text: "This is a label"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  | -------- This is a label
`, got)
}

func TestRender_MultipleLabelsOnOneLine(t *testing.T) {
	r := New()
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerTilde, Text: "This is foo"},
		{Span: Span{Start: 0, End: 8}, Marker: MarkerDash, Text: "This is the whole line"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  | ~~~ This is foo
  | -------- This is the whole line
`, got)
}

func TestRender_MultipleLabelsOnOneLine_Unsorted(t *testing.T) {
	r := New()
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 0, End: 8}, Marker: MarkerDash, Text: "whole"},
		{Span: Span{Start: 0, End: 3}, Marker: MarkerTilde, Text: "foo"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  | ~~~ foo
  | -------- whole
`, got)
}

func TestRender_MultiLineSpan(t *testing.T) {
	r := New()
	src := []byte("foo: bar\nbaz: qux")
	labels := []Label{
		{Span: Span{Start: 0, End: 17}, Marker: MarkerDash, Text: "This is a label"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  | --------
2 | baz: qux
  | -------- This is a label
`, got)
}

func TestRender_MultiLineSpanPartialFirstLine(t *testing.T) {
	r := New()
	src := []byte("foo: bar\nbaz: qux")
	labels := []Label{
		{Span: Span{Start: 5, End: 17}, Marker: MarkerDash, Text: "This is a label"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  |      ---
2 | baz: qux
  | -------- This is a label
`, got)
}

func TestRender_MultiLineSpanPartialLastLine(t *testing.T) {
	r := New()
	src := []byte("foo: bar\nbaz: qux")
	labels := []Label{
		{Span: Span{Start: 0, End: 12}, Marker: MarkerDash, Text: "partial"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  | --------
2 | baz: qux
  | --- partial
`, got)
}

func TestRender_MultiLineSpanThreeLines(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 1, End: 10}, Marker: MarkerDash, Text: "spans three lines"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | aaa
  |  --
2 | bbb
  | ---
3 | ccc
  | -- spans three lines
`, got)
}

func TestRender_TabExpansion(t *testing.T) {
	r := New()
	src := []byte("\tfoo")
	labels := []Label{
		{Span: Span{Start: 0, End: 4}, Marker: MarkerDash, Text: "label"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 |     foo
  | ------- label
`, got)
}

func TestRender_TabExpansion_LabelAfterTab(t *testing.T) {
	r := New()
	src := []byte("\tfoo")
	labels := []Label{
		{Span: Span{Start: 1, End: 4}, Marker: MarkerDash, Text: "after tab"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 |     foo
  |     --- after tab
`, got)
}

func TestRender_WideCharacters(t *testing.T) {
	r := New()
	src := []byte("こんにちは")
	labels := []Label{
		{Span: Span{Start: 0, End: 15}, Marker: MarkerDash, Text: "greeting"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | こんにちは
  | ---------- greeting
`, got)
}

func TestRender_WideCharacters_PartialSpan(t *testing.T) {
	r := New()
	src := []byte("こんにちは")
	labels := []Label{
		{Span: Span{Start: 6, End: 15}, Marker: MarkerTilde, Text: "last three"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | こんにちは
  |     ~~~~~~ last three
`, got)
}

func TestRender_InvalidSpan(t *testing.T) {
	r := New()
	src := []byte("foo")

	tests := []struct {
		name   string
		labels []Label
	}{
		{"start < 0", []Label{{Span: Span{Start: -1, End: 3}, Marker: MarkerDash}}},
		{"end > len(src)", []Label{{Span: Span{Start: 0, End: 10}, Marker: MarkerDash}}},
		{"start == end", []Label{{Span: Span{Start: 3, End: 3}, Marker: MarkerDash}}},
		{"start > end", []Label{{Span: Span{Start: 3, End: 1}, Marker: MarkerDash}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Render(src, tt.labels)
			assert.Error(t, err)
		})
	}
}

func TestRender_TrailingNewline(t *testing.T) {
	r := New()

	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{"single line with newline", "foo\n", ""},
		{"two lines with double newline", "foo\n\n", ""},
		{"just a newline", "\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Render([]byte(tt.src), nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRender_SpecExample(t *testing.T) {
	r := New()
	src := []byte("foo: bar\nbaz: qux\n")
	labels := []Label{
		{Span: Span{Start: 0, End: 8}, Marker: MarkerDash, Text: "This is a label for the first line"},
		{Span: Span{Start: 9, End: 17}, Marker: MarkerTilde, Text: "This is a label for the second line"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo: bar
  | -------- This is a label for the first line
2 | baz: qux
  | ~~~~~~~~ This is a label for the second line
`, got)
}

func TestRender_EmptyLabelText(t *testing.T) {
	r := New()
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: ""},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo
  | ---
`, got)
}

func TestWrite(t *testing.T) {
	r := New()
	src := []byte("hello")
	labels := []Label{
		{Span: Span{Start: 0, End: 5}, Marker: MarkerCaret, Text: "greeting"},
	}
	var buf strings.Builder
	err := r.Write(&buf, src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | hello
  | ^^^^^ greeting
`, buf.String())
}

func TestRender_OverlappingSpans(t *testing.T) {
	r := New()
	src := []byte("foobarbaz")
	labels := []Label{
		{Span: Span{Start: 0, End: 6}, Marker: MarkerDash, Text: "first"},
		{Span: Span{Start: 3, End: 9}, Marker: MarkerTilde, Text: "second"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foobarbaz
  | ------ first
  |    ~~~~~~ second
`, got)
}

func TestRender_MultiDigitLineNumbers(t *testing.T) {
	r := New()
	src := []byte("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	got, err := r.Render(src, nil)
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestRender_MultiDigitLineNumbers_WithLabel(t *testing.T) {
	r := New()
	src := []byte("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	labels := []Label{
		{Span: Span{Start: 18, End: 20}, Marker: MarkerDash, Text: "last line"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n10 | 10\n   | -- last line\n", got)
}

func TestRender_ArbitraryMarker(t *testing.T) {
	r := New()
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: LabelMarker('*'), Text: "star"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo
  | *** star
`, got)
}

func TestRender_DefaultMarker(t *testing.T) {
	r := New()
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Text: "no marker specified"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo
  | --- no marker specified
`, got)
}

func TestRender_MarkerNone_WithText(t *testing.T) {
	r := New()
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 5, End: 8}, Marker: MarkerNone, Text: "value"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo: bar\n  |      value\n", got)
}

func TestRender_MarkerNone_EmptyText(t *testing.T) {
	r := New()
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 5, End: 8}, Marker: MarkerNone, Text: ""},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo: bar\n", got)
}

func TestRender_MarkerNone_MultiLineSpan(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 1, End: 10}, Marker: MarkerNone, Text: "spans three lines"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | aaa\n2 | bbb\n3 | ccc\n  | spans three lines\n", got)
}

func TestRender_MarkerNone_MultiLineSpan_EmptyText(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 1, End: 10}, Marker: MarkerNone, Text: ""},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | aaa\n2 | bbb\n3 | ccc\n", got)
}

func TestRender_MarkerNone_MixedWithOtherMarkers(t *testing.T) {
	r := New()
	src := []byte("foo bar baz")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "dashed"},
		{Span: Span{Start: 4, End: 7}, Marker: MarkerNone, Text: "no marker"},
		{Span: Span{Start: 8, End: 11}, Marker: MarkerTilde, Text: "tilde"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo bar baz\n  | --- dashed\n  |     no marker\n  |         ~~~ tilde\n", got)
}

func TestRender_MarkerNone_WithLabelTextStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })
	r := New()
	r.Style = Style{LabelText: bracket}
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerNone, Text: "note"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo\n  | [note]\n", got)
}

func TestRender_MarkerNone_WithLabelStyleOverride(t *testing.T) {
	global := StyleFunc(func(s string) string { return "G(" + s + ")" })
	local := StyleFunc(func(s string) string { return "L(" + s + ")" })

	r := New()
	r.Style = Style{LabelText: global}
	src := []byte("foo")
	labels := []Label{
		{
			Span: Span{Start: 0, End: 3}, Marker: MarkerNone, Text: "note",
			Style: LabelStyle{LabelText: local},
		},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo\n  | L(note)\n", got)
}

func TestRender_MarkerLineStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	r := New()
	r.Style = Style{
		Marker:    bracket,
		LabelText: bracket,
	}
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "label"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo\n  | [---] [label]\n", got)
}

func TestRender_MarkerLineStyle_LabelOverride(t *testing.T) {
	global := StyleFunc(func(s string) string { return "G(" + s + ")" })
	local := StyleFunc(func(s string) string { return "L(" + s + ")" })

	r := New()
	r.Style = Style{Marker: global, LabelText: global}
	src := []byte("foo")
	labels := []Label{
		{
			Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "label",
			Style: LabelStyle{Marker: local, LabelText: local},
		},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | foo\n  | L(---) L(label)\n", got)
}

func TestRender_CodeLineStyle_LineNumberAndSeparator(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	r := New()
	r.Style = Style{
		LineNumber: bracket,
		Separator:  bracket,
	}
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "[1] [|] foo\n  [|] --- x\n", got)
}

func TestRender_CodeLineStyle_LabelOverridesLineNumber(t *testing.T) {
	global := StyleFunc(func(s string) string { return "G(" + s + ")" })
	local := StyleFunc(func(s string) string { return "L(" + s + ")" })

	r := New()
	r.Style = Style{LineNumber: global, Separator: global}
	src := []byte("foo")
	labels := []Label{
		{
			Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x",
			Style: LabelStyle{LineNumber: local, Separator: local},
		},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "L(1) L(|) foo\n  L(|) --- x\n", got)
}

func TestRender_NoLabelsEmptyOutput(t *testing.T) {
	r := New()
	src := []byte("foo\nbar\nbaz")
	got, err := r.Render(src, nil)
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestRender_SpanCodeStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	r := New()
	r.Style = Style{SpanCode: bracket}
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 5, End: 8}, Marker: MarkerDash, Text: "val"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// "foo: " は非span、"bar" は span
	assert.Equal(t, "1 | foo: [bar]\n  |      --- val\n", got)
}

func TestRender_NonSpanCodeStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	r := New()
	r.Style = Style{NonSpanCode: bracket}
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 5, End: 8}, Marker: MarkerDash, Text: "val"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | [foo: ]bar\n  |      --- val\n", got)
}

func TestRender_NoLabelLine_NonSpanCodeStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	r := New(WithAfter(1))
	r.Style = Style{NonSpanCode: bracket}
	src := []byte("foo\nbar")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// Line 1 has label so "foo" is SpanCode (no SpanCode style set, passthrough).
	// Line 2 is context (no label), so "bar" gets NonSpanCode style.
	assert.Equal(t, "1 | foo\n  | --- x\n2 | [bar]\n", got)
}

func TestRender_SpanCodeStyle_MultipleLabels(t *testing.T) {
	red := StyleFunc(func(s string) string { return "R(" + s + ")" })
	blue := StyleFunc(func(s string) string { return "B(" + s + ")" })

	r := New()
	src := []byte("foobarbaz")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "a", Style: LabelStyle{SpanCode: red}},
		{Span: Span{Start: 6, End: 9}, Marker: MarkerTilde, Text: "b", Style: LabelStyle{SpanCode: blue}},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// "foo" = red, "bar" = 非span, "baz" = blue
	assert.Equal(t, "1 | R(foo)barB(baz)\n  | --- a\n  |       ~~~ b\n", got)
}

func TestRender_SpanCodeStyle_OverlappingLabels(t *testing.T) {
	red := StyleFunc(func(s string) string { return "R(" + s + ")" })
	blue := StyleFunc(func(s string) string { return "B(" + s + ")" })

	r := New()
	src := []byte("foobar")
	labels := []Label{
		{Span: Span{Start: 0, End: 6}, Marker: MarkerDash, Text: "all", Style: LabelStyle{SpanCode: red}},
		{Span: Span{Start: 0, End: 3}, Marker: MarkerTilde, Text: "first", Style: LabelStyle{SpanCode: blue}},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// ソート順: Span{0,3} (短い) が先 → "foo" = blue, "bar" = red
	assert.Equal(t, "1 | B(foo)R(bar)\n  | ~~~ first\n  | ------ all\n", got)
}

func TestRender_FullStyleIntegration(t *testing.T) {
	r := New()
	r.Style = Style{
		LineNumber:  StyleFunc(func(s string) string { return "<ln:" + s + ">" }),
		Separator:   StyleFunc(func(s string) string { return "<sep:" + s + ">" }),
		SpanCode:    StyleFunc(func(s string) string { return "<span:" + s + ">" }),
		NonSpanCode: StyleFunc(func(s string) string { return "<non:" + s + ">" }),
		Marker:      StyleFunc(func(s string) string { return "<mark:" + s + ">" }),
		LabelText:   StyleFunc(func(s string) string { return "<text:" + s + ">" }),
	}

	src := []byte("foo: bar\nbaz: qux")
	labels := []Label{
		{Span: Span{Start: 5, End: 8}, Marker: MarkerDash, Text: "value"},
	}

	got, err := r.Render(src, labels)
	require.NoError(t, err)

	// Line 1 has a label on "bar" (bytes 5-8)
	// Line 2 has no label and is filtered out
	expected := "" +
		"<ln:1> <sep:|> <non:foo: ><span:bar>\n" +
		"  <sep:|>      <mark:---> <text:value>\n" +
		"...\n"
	assert.Equal(t, expected, got)
}

func TestRender_StyleIntegration_LabelOverrideAndFallback(t *testing.T) {
	global := StyleFunc(func(s string) string { return "G(" + s + ")" })
	local := StyleFunc(func(s string) string { return "L(" + s + ")" })

	r := New()
	r.Style = Style{
		LineNumber: global,
		Separator:  global,
		SpanCode:   global,
		Marker:     global,
		LabelText:  global,
	}

	src := []byte("foo bar")
	labels := []Label{
		{
			Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "overridden",
			Style: LabelStyle{
				LineNumber: local, Separator: local,
				SpanCode: local, Marker: local, LabelText: local,
			},
		},
		{
			Span: Span{Start: 4, End: 7}, Marker: MarkerTilde, Text: "global fallback",
			// Style 未設定 → グローバルにフォールバック
		},
	}

	got, err := r.Render(src, labels)
	require.NoError(t, err)

	// First label (startInLine=0) overrides line number and separator for the code line.
	// "foo" uses local SpanCode, " " is non-span (no NonSpanCode set so passthrough), "bar" uses global SpanCode.
	// First marker line uses local styles, second uses global styles.
	expected := "" +
		"L(1) L(|) L(foo) G(bar)\n" +
		"  L(|) L(---) L(overridden)\n" +
		"  G(|)     G(~~~) G(global fallback)\n"
	assert.Equal(t, expected, got)
}

func TestRender_StyleIntegration_MultiLineSpan(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	r := New()
	r.Style = Style{SpanCode: bracket, Marker: bracket}

	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 1, End: 10}, Marker: MarkerDash, Text: "spans three"},
	}

	got, err := r.Render(src, labels)
	require.NoError(t, err)

	// Span covers bytes 1-10: "aa" on line 1, all of "bbb" on line 2, "cc" on line 3
	expected := "" +
		"1 | a[aa]\n" +
		"  |  [--]\n" +
		"2 | [bbb]\n" +
		"  | [---]\n" +
		"3 | [cc]c\n" +
		"  | [--] spans three\n"
	assert.Equal(t, expected, got)
}

func TestRender_FilterLabelsOnly(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "first"},
		{Span: Span{Start: 16, End: 19}, Marker: MarkerTilde, Text: "last"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | aaa\n  | --- first\n...\n5 | eee\n  | ~~~ last\n", got)
}

func TestRender_FilterWithBefore(t *testing.T) {
	r := New(WithBefore(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "middle"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n2 | bbb\n3 | ccc\n  | --- middle\n...\n", got)
}

func TestRender_FilterWithAfter(t *testing.T) {
	r := New(WithAfter(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "middle"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n3 | ccc\n  | --- middle\n4 | ddd\n...\n", got)
}

func TestRender_FilterWithBeforeAndAfter(t *testing.T) {
	r := New(WithBefore(1), WithAfter(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "middle"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n2 | bbb\n3 | ccc\n  | --- middle\n4 | ddd\n...\n", got)
}

func TestRender_FilterPerLabelOverride(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "with context", Before: new(1)},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n2 | bbb\n3 | ccc\n  | --- with context\n...\n", got)
}

func TestRender_FilterPerLabelOverrideZero(t *testing.T) {
	r := New(WithBefore(2))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "no context", Before: new(0)},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n3 | ccc\n  | --- no context\n...\n", got)
}

func TestRender_FilterEllipsisWideLineNumbers(t *testing.T) {
	r := New()
	var src []byte
	for i := 1; i <= 100; i++ {
		if i > 1 {
			src = append(src, '\n')
		}
		src = append(src, fmt.Appendf(nil, "line%d", i)...)
	}
	labels := []Label{
		{Span: Span{Start: 0, End: 5}, Marker: MarkerDash, Text: "first"},
		{Span: Span{Start: len(src) - 7, End: len(src)}, Marker: MarkerTilde, Text: "last"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// Line number width is 3 (max visible is 100). "..." is exactly 3 chars.
	expected := "" +
		"  1 | line1\n" +
		"    | ----- first\n" +
		"...\n" +
		"100 | line100\n" +
		"    | ~~~~~~~ last\n"
	assert.Equal(t, expected, got)
}

func TestRender_FilterEllipsisStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })
	r := New(WithStyle(Style{Ellipsis: bracket}))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "first"},
		{Span: Span{Start: 16, End: 19}, Marker: MarkerTilde, Text: "last"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | aaa\n  | --- first\n[...]\n5 | eee\n  | ~~~ last\n", got)
}

func TestRender_FilterEllipsisStyle_LeadingTrailing(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })
	r := New(WithStyle(Style{Ellipsis: bracket}))
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 4, End: 7}, Marker: MarkerDash, Text: "middle"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "[...]\n2 | bbb\n  | --- middle\n[...]\n", got)
}

func TestRender_FilterAdjacentLabels(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "first"},
		{Span: Span{Start: 4, End: 7}, Marker: MarkerTilde, Text: "second"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | aaa\n  | --- first\n2 | bbb\n  | ~~~ second\n...\n", got)
}

func TestRender_FilterOverlappingContext(t *testing.T) {
	r := New(WithBefore(1), WithAfter(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 4, End: 7}, Marker: MarkerDash, Text: "second"},
		{Span: Span{Start: 12, End: 15}, Marker: MarkerTilde, Text: "fourth"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | aaa\n2 | bbb\n  | --- second\n3 | ccc\n4 | ddd\n  | ~~~ fourth\n5 | eee\n", got)
}

func TestRender_FilterLeadingTrailingEllipsis(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "middle"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n3 | ccc\n  | --- middle\n...\n", got)
}

func TestRender_FilterLineNumberWidth(t *testing.T) {
	r := New()
	src := []byte("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12")
	labels := []Label{
		{Span: Span{Start: 2, End: 3}, Marker: MarkerDash, Text: "line 2"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n2 | 2\n  | - line 2\n...\n", got)
}

func TestRender_FilterMultiLineSpanWithContext(t *testing.T) {
	r := New(WithBefore(1), WithAfter(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee\nfff")
	// Label spans lines 3-4 (bytes 8-15: "ccc\nddd")
	labels := []Label{
		{Span: Span{Start: 8, End: 15}, Marker: MarkerDash, Text: "two lines"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// Before=1 applies to line 3 → line 2 visible
	// After=1 applies to line 4 → line 5 visible
	expected := "" +
		"...\n" +
		"2 | bbb\n" +
		"3 | ccc\n" +
		"  | ---\n" +
		"4 | ddd\n" +
		"  | --- two lines\n" +
		"5 | eee\n" +
		"...\n"
	assert.Equal(t, expected, got)
}

func TestRender_FilterPerLabelNegativeClamped(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 4, End: 7}, Marker: MarkerDash, Text: "middle", Before: new(-3), After: new(-5)},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// Negative values clamped to 0, so only the labeled line is shown
	assert.Equal(t, "...\n2 | bbb\n  | --- middle\n...\n", got)
}

func TestRender_FilterPerLabelAfterOverride(t *testing.T) {
	r := New()
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 4, End: 7}, Marker: MarkerDash, Text: "with after", After: new(2)},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "...\n2 | bbb\n  | --- with after\n3 | ccc\n4 | ddd\n...\n", got)
}

func TestRender_FilterMixedOverrides(t *testing.T) {
	r := New(WithBefore(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 4, End: 7}, Marker: MarkerDash, Text: "override", Before: new(0)},
		{Span: Span{Start: 12, End: 15}, Marker: MarkerTilde, Text: "default"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	// First label: Before overridden to 0, only line 2
	// Second label: Before=1 from renderer default, lines 3 and 4
	expected := "" +
		"...\n" +
		"2 | bbb\n" +
		"  | --- override\n" +
		"3 | ccc\n" +
		"4 | ddd\n" +
		"  | ~~~ default\n" +
		"...\n"
	assert.Equal(t, expected, got)
}

func TestRender_SourceStyle(t *testing.T) {
	styler := func(src string) string {
		var lines []string
		for _, line := range strings.Split(src, "\n") {
			lines = append(lines, "\033[32m"+line+"\033[0m")
		}
		return strings.Join(lines, "\n")
	}

	r := New(WithSourceStyle(styler))
	src := []byte("foo: bar")
	labels := []Label{
		{Span: Span{Start: 0, End: 8}, Marker: MarkerDash, Text: "whole line"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | \033[32mfoo: bar\033[0m\n  | -------- whole line\n", got)
}

func TestRender_SourceStyle_MultiLine(t *testing.T) {
	styler := func(src string) string {
		var lines []string
		for _, line := range strings.Split(src, "\n") {
			lines = append(lines, "\033[33m"+line+"\033[0m")
		}
		return strings.Join(lines, "\n")
	}

	r := New(WithSourceStyle(styler))
	src := []byte("foo\nbar\nbaz")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "first"},
		{Span: Span{Start: 8, End: 11}, Marker: MarkerTilde, Text: "last"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	expected := "" +
		"1 | \033[33mfoo\033[0m\n" +
		"  | --- first\n" +
		"...\n" +
		"3 | \033[33mbaz\033[0m\n" +
		"  | ~~~ last\n"
	assert.Equal(t, expected, got)
}

func TestRender_SourceStyle_IgnoresSpanCodeStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })
	styler := func(src string) string {
		return "\033[32m" + src + "\033[0m"
	}

	r := New(
		WithSourceStyle(styler),
		WithStyle(Style{
			SpanCode:    bracket,
			NonSpanCode: bracket,
			Marker:      bracket,
		}),
	)
	src := []byte("foo")
	labels := []Label{
		{
			Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x",
			Style: LabelStyle{SpanCode: bracket},
		},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | \033[32mfoo\033[0m\n  | [---] x\n", got)
}

func TestRender_SourceStyle_TrailingNewline(t *testing.T) {
	styler := func(src string) string {
		return "\033[32m" + src + "\033[0m\n"
	}

	r := New(WithSourceStyle(styler))
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | \033[32mfoo\033[0m\n  | --- x\n", got)
}

func TestRender_SourceStyle_LineCountMismatch(t *testing.T) {
	styler := func(src string) string {
		return src + "\nextra line"
	}

	r := New(WithSourceStyle(styler))
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x"},
	}
	_, err := r.Render(src, labels)
	assert.Error(t, err)
}

func TestRender_SourceStyle_TabExpansion(t *testing.T) {
	var received string
	styler := func(src string) string {
		received = src
		return "\033[32m" + src + "\033[0m"
	}

	r := New(WithSourceStyle(styler))
	src := []byte("\tfoo")
	labels := []Label{
		{Span: Span{Start: 0, End: 4}, Marker: MarkerDash, Text: "label"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "    foo", received)
	assert.Equal(t, "1 | \033[32m    foo\033[0m\n  | ------- label\n", got)
}

func TestRender_SourceStyle_WithContext(t *testing.T) {
	styler := func(src string) string {
		var lines []string
		for _, line := range strings.Split(src, "\n") {
			lines = append(lines, "\033[32m"+line+"\033[0m")
		}
		return strings.Join(lines, "\n")
	}

	r := New(WithSourceStyle(styler), WithBefore(1), WithAfter(1))
	src := []byte("aaa\nbbb\nccc\nddd\neee")
	labels := []Label{
		{Span: Span{Start: 8, End: 11}, Marker: MarkerDash, Text: "middle"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	expected := "" +
		"...\n" +
		"2 | \033[32mbbb\033[0m\n" +
		"3 | \033[32mccc\033[0m\n" +
		"  | --- middle\n" +
		"4 | \033[32mddd\033[0m\n" +
		"...\n"
	assert.Equal(t, expected, got)
}

func TestRender_SourceStyle_MultiLineSpan(t *testing.T) {
	styler := func(src string) string {
		var lines []string
		for _, line := range strings.Split(src, "\n") {
			lines = append(lines, "\033[33m"+line+"\033[0m")
		}
		return strings.Join(lines, "\n")
	}

	r := New(WithSourceStyle(styler))
	src := []byte("aaa\nbbb\nccc")
	labels := []Label{
		{Span: Span{Start: 1, End: 10}, Marker: MarkerDash, Text: "spans three lines"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	expected := "" +
		"1 | \033[33maaa\033[0m\n" +
		"  |  --\n" +
		"2 | \033[33mbbb\033[0m\n" +
		"  | ---\n" +
		"3 | \033[33mccc\033[0m\n" +
		"  | -- spans three lines\n"
	assert.Equal(t, expected, got)
}

func TestRender_SourceStyle_MultipleTrailingNewlines(t *testing.T) {
	styler := func(src string) string {
		return "\033[32m" + src + "\033[0m\n\n\n"
	}

	r := New(WithSourceStyle(styler))
	src := []byte("foo")
	labels := []Label{
		{Span: Span{Start: 0, End: 3}, Marker: MarkerDash, Text: "x"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, "1 | \033[32mfoo\033[0m\n  | --- x\n", got)
}
