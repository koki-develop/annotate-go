package annotate

import (
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
	assert.Equal(t, `1 | foo: bar
`, got)
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
		{
			"single line with newline",
			"foo\n",
			`1 | foo
`,
		},
		{
			"two lines with double newline",
			"foo\n\n",
			"1 | foo\n" +
				"2 | \n",
		},
		{
			"just a newline",
			"\n",
			"1 | \n",
		},
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
	assert.Equal(t, ` 1 | 1
 2 | 2
 3 | 3
 4 | 4
 5 | 5
 6 | 6
 7 | 7
 8 | 8
 9 | 9
10 | 10
`, got)
}

func TestRender_MultiDigitLineNumbers_WithLabel(t *testing.T) {
	r := New()
	src := []byte("1\n2\n3\n4\n5\n6\n7\n8\n9\n10")
	labels := []Label{
		{Span: Span{Start: 18, End: 20}, Marker: MarkerDash, Text: "last line"},
	}
	got, err := r.Render(src, labels)
	require.NoError(t, err)
	assert.Equal(t, ` 1 | 1
 2 | 2
 3 | 3
 4 | 4
 5 | 5
 6 | 6
 7 | 7
 8 | 8
 9 | 9
10 | 10
   | -- last line
`, got)
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

func TestApplyStyle(t *testing.T) {
	bracket := StyleFunc(func(s string) string { return "[" + s + "]" })

	tests := []struct {
		name     string
		s        string
		fn       StyleFunc
		expected string
	}{
		{"nil func passthrough", "hello", nil, "hello"},
		{"applies func", "hello", bracket, "[hello]"},
		{"empty string skipped", "", bracket, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, applyStyle(tt.s, tt.fn))
		})
	}
}

func TestApplyStyle_EmptyStringDoesNotCallFunc(t *testing.T) {
	called := false
	fn := StyleFunc(func(s string) string {
		called = true
		return s
	})
	applyStyle("", fn)
	assert.False(t, called)
}

func TestResolveStyle(t *testing.T) {
	global := StyleFunc(func(s string) string { return "G(" + s + ")" })
	label := StyleFunc(func(s string) string { return "L(" + s + ")" })

	tests := []struct {
		name        string
		labelStyle  StyleFunc
		globalStyle StyleFunc
		input       string
		expected    string
	}{
		{"label wins over global", label, global, "x", "L(x)"},
		{"global used when label nil", nil, global, "x", "G(x)"},
		{"passthrough when both nil", nil, nil, "x", "x"},
		{"empty string skipped", label, global, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, resolveStyle(tt.input, tt.labelStyle, tt.globalStyle))
		})
	}
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
	got, err := r.Render(src, nil)
	require.NoError(t, err)
	assert.Equal(t, "[1] [|] foo\n", got)
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

func TestRender_MultipleLinesNoLabels(t *testing.T) {
	r := New()
	src := []byte("foo\nbar\nbaz")
	got, err := r.Render(src, nil)
	require.NoError(t, err)
	assert.Equal(t, `1 | foo
2 | bar
3 | baz
`, got)
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

	r := New()
	r.Style = Style{NonSpanCode: bracket}
	src := []byte("foo")
	got, err := r.Render(src, nil)
	require.NoError(t, err)
	assert.Equal(t, "1 | [foo]\n", got)
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
