package annotate

// StyleFunc is a function that transforms a string for visual styling.
// It is typically used to wrap text with ANSI escape codes or other formatting.
// A nil StyleFunc applies no styling.
type StyleFunc func(string) string

// Style controls the visual styling applied to each element of the rendered output.
// Each field is a [StyleFunc] applied to the corresponding element.
// Nil fields leave the element unstyled.
type Style struct {
	// LineNumber styles the line number column (e.g., "1", "2", "10").
	LineNumber StyleFunc
	// Separator styles the "|" separator between the line number and code.
	Separator StyleFunc
	// SpanCode styles the portion of source code covered by a label's span.
	SpanCode StyleFunc
	// NonSpanCode styles the portion of source code not covered by any label's span.
	NonSpanCode StyleFunc
	// Marker styles the marker characters (e.g., "---", "~~~") on annotation lines.
	Marker StyleFunc
	// LabelText styles the descriptive text on annotation lines.
	LabelText StyleFunc
}

// LabelStyle provides per-label style overrides. Non-nil fields take precedence
// over the corresponding fields in the global [Style] for this label only.
//
// Unlike [Style], LabelStyle does not have a NonSpanCode field because
// non-span code is not associated with any particular label.
type LabelStyle struct {
	// LineNumber overrides [Style.LineNumber] for lines containing this label.
	LineNumber StyleFunc
	// Separator overrides [Style.Separator] for lines containing this label.
	Separator StyleFunc
	// SpanCode overrides [Style.SpanCode] for the span covered by this label.
	SpanCode StyleFunc
	// Marker overrides [Style.Marker] for this label's marker characters.
	Marker StyleFunc
	// LabelText overrides [Style.LabelText] for this label's descriptive text.
	LabelText StyleFunc
}

func applyStyle(s string, fn StyleFunc) string {
	if fn == nil || s == "" {
		return s
	}
	return fn(s)
}

func resolveStyle(s string, labelStyle, globalStyle StyleFunc) string {
	if labelStyle != nil {
		return applyStyle(s, labelStyle)
	}
	return applyStyle(s, globalStyle)
}

func ansiStyle(code string) StyleFunc {
	return func(s string) string {
		return "\033[" + code + "m" + s + "\033[0m"
	}
}

// Built-in ANSI text attribute [StyleFunc] variables.
var (
	Bold      = ansiStyle("1")
	Dim       = ansiStyle("2")
	Italic    = ansiStyle("3")
	Underline = ansiStyle("4")
)

// Built-in ANSI foreground color [StyleFunc] variables.
var (
	FgRed     = ansiStyle("31")
	FgGreen   = ansiStyle("32")
	FgYellow  = ansiStyle("33")
	FgBlue    = ansiStyle("34")
	FgMagenta = ansiStyle("35")
	FgCyan    = ansiStyle("36")
	FgWhite   = ansiStyle("37")
)

// Built-in ANSI background color [StyleFunc] variables.
var (
	BgRed     = ansiStyle("41")
	BgGreen   = ansiStyle("42")
	BgYellow  = ansiStyle("43")
	BgBlue    = ansiStyle("44")
	BgMagenta = ansiStyle("45")
	BgCyan    = ansiStyle("46")
	BgWhite   = ansiStyle("47")
)

// DefaultStyle is a [Style] preset that dims line numbers and separators.
var DefaultStyle = Style{
	LineNumber: Dim,
	Separator:  Dim,
}

// LabelStyleError is a [LabelStyle] preset for error diagnostics.
// Markers are red; label text is red and bold.
var LabelStyleError = LabelStyle{
	Marker:    FgRed,
	LabelText: ComposeStyles(FgRed, Bold),
}

// LabelStyleWarning is a [LabelStyle] preset for warning diagnostics.
// Markers and label text are yellow.
var LabelStyleWarning = LabelStyle{
	Marker:    FgYellow,
	LabelText: FgYellow,
}

// LabelStyleInfo is a [LabelStyle] preset for informational diagnostics.
// Markers and label text are blue.
var LabelStyleInfo = LabelStyle{
	Marker:    FgBlue,
	LabelText: FgBlue,
}

// LabelStyleHint is a [LabelStyle] preset for hint diagnostics.
// Markers and label text are dimmed.
var LabelStyleHint = LabelStyle{
	Marker:    Dim,
	LabelText: Dim,
}

// ComposeStyles combines multiple [StyleFunc] functions into one.
// The functions are applied in the order given (left to right),
// so the first function becomes the outermost wrapper.
//
//	redBold := ComposeStyles(FgRed, Bold)
//	redBold("text") // equivalent to FgRed(Bold("text"))
func ComposeStyles(fns ...StyleFunc) StyleFunc {
	return func(s string) string {
		for i := len(fns) - 1; i >= 0; i-- {
			s = fns[i](s)
		}
		return s
	}
}
