package annotate

type StyleFunc func(string) string

type Style struct {
	LineNumber  StyleFunc
	Separator   StyleFunc
	SpanCode    StyleFunc
	NonSpanCode StyleFunc
	Marker      StyleFunc
	LabelText   StyleFunc
}

type LabelStyle struct {
	LineNumber StyleFunc
	Separator  StyleFunc
	SpanCode   StyleFunc
	Marker     StyleFunc
	LabelText  StyleFunc
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

var (
	Bold      = ansiStyle("1")
	Dim       = ansiStyle("2")
	Italic    = ansiStyle("3")
	Underline = ansiStyle("4")
)

var (
	FgRed     = ansiStyle("31")
	FgGreen   = ansiStyle("32")
	FgYellow  = ansiStyle("33")
	FgBlue    = ansiStyle("34")
	FgMagenta = ansiStyle("35")
	FgCyan    = ansiStyle("36")
	FgWhite   = ansiStyle("37")
)

var (
	BgRed     = ansiStyle("41")
	BgGreen   = ansiStyle("42")
	BgYellow  = ansiStyle("43")
	BgBlue    = ansiStyle("44")
	BgMagenta = ansiStyle("45")
	BgCyan    = ansiStyle("46")
	BgWhite   = ansiStyle("47")
)

var DefaultStyle = Style{
	LineNumber: Dim,
	Separator:  Dim,
}

var LabelStyleError = LabelStyle{
	Marker:    FgRed,
	LabelText: ComposeStyles(FgRed, Bold),
}

var LabelStyleWarning = LabelStyle{
	Marker:    FgYellow,
	LabelText: FgYellow,
}

var LabelStyleInfo = LabelStyle{
	Marker:    FgBlue,
	LabelText: FgBlue,
}

var LabelStyleHint = LabelStyle{
	Marker:    Dim,
	LabelText: Dim,
}

func ComposeStyles(fns ...StyleFunc) StyleFunc {
	return func(s string) string {
		for i := len(fns) - 1; i >= 0; i-- {
			s = fns[i](s)
		}
		return s
	}
}
