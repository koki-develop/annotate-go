package annotate

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

type Span struct {
	Start int
	End   int
}

type LabelMarker rune

const (
	MarkerDash  LabelMarker = '-'
	MarkerTilde LabelMarker = '~'
	MarkerCaret LabelMarker = '^'
)

type Label struct {
	Span   Span
	Marker LabelMarker
	Text   string
	Style  LabelStyle
}

type Renderer struct {
	Style Style
}

func New(opts ...Option) *Renderer {
	r := &Renderer{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Renderer) Render(src []byte, labels []Label) (string, error) {
	var buf strings.Builder
	if err := r.Write(&buf, src, labels); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type line struct {
	number    int
	text      string // tab-expanded text
	rawLen    int    // byte length of original text (before tab expansion)
	start     int    // byte offset in source
	end       int    // byte offset in source (exclusive)
	offsetMap []int  // maps original byte offset (relative to line) to expanded byte offset
}

const tabWidth = 4

func expandTabs(s string) (string, []int) {
	var buf strings.Builder
	offsetMap := make([]int, len(s)+1)
	col := 0
	for i, ch := range s {
		expandedPos := buf.Len()
		offsetMap[i] = expandedPos
		runeLen := utf8.RuneLen(ch)
		for j := 1; j < runeLen; j++ {
			offsetMap[i+j] = expandedPos
		}
		if ch == '\t' {
			spaces := tabWidth - (col % tabWidth)
			buf.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		} else {
			buf.WriteRune(ch)
			col += runewidth.RuneWidth(ch)
		}
	}
	offsetMap[len(s)] = buf.Len()
	return buf.String(), offsetMap
}

func parseLines(src []byte) []line {
	if len(src) == 0 {
		return nil
	}

	raw := strings.Split(string(src), "\n")
	if len(raw) > 1 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}

	lines := make([]line, len(raw))
	offset := 0
	for i, text := range raw {
		end := offset + len(text)
		expanded, offsetMap := expandTabs(text)
		lines[i] = line{
			number:    i + 1,
			text:      expanded,
			rawLen:    len(text),
			start:     offset,
			end:       end,
			offsetMap: offsetMap,
		}
		offset = end + 1
	}
	return lines
}

type lineLabel struct {
	label       *Label
	startInLine int
	endInLine   int
	isLastLine  bool
}

func mapLabelsToLines(lines []line, labels []Label) map[int][]lineLabel {
	result := make(map[int][]lineLabel)
	for i := range labels {
		lbl := &labels[i]
		for li, ln := range lines {
			spanStart := lbl.Span.Start
			spanEnd := lbl.Span.End
			lineStart := ln.start
			lineEnd := ln.end

			if spanStart >= lineEnd+1 || spanEnd <= lineStart {
				continue
			}

			startInLine := 0
			if spanStart > lineStart {
				startInLine = spanStart - lineStart
			}
			endInLine := ln.rawLen
			if spanEnd < lineEnd+1 {
				endInLine = spanEnd - lineStart
			}

			isLast := true
			if li+1 < len(lines) && spanEnd > lineEnd+1 {
				isLast = false
			}

			result[li] = append(result[li], lineLabel{
				label:       lbl,
				startInLine: startInLine,
				endInLine:   endInLine,
				isLastLine:  isLast,
			})
		}
	}
	return result
}

func segmentCodeLine(ln line, ll []lineLabel, globalSpanCode, globalNonSpanCode StyleFunc) string {
	if len(ll) == 0 {
		return applyStyle(ln.text, globalNonSpanCode)
	}

	// Build span ranges from labels (already sorted by caller)
	type spanRange struct {
		start int // expanded byte offset
		end   int // expanded byte offset
		style StyleFunc
	}
	var ranges []spanRange
	for _, lbl := range ll {
		expandedStart := ln.offsetMap[lbl.startInLine]
		expandedEnd := ln.offsetMap[lbl.endInLine]
		ranges = append(ranges, spanRange{
			start: expandedStart,
			end:   expandedEnd,
			style: lbl.label.Style.SpanCode,
		})
	}

	// Remove overlaps (first wins).
	// Precondition: ll is sorted by startInLine ascending, so ranges are also ordered by start.
	var merged []spanRange
	for _, r := range ranges {
		clipped := r
		for _, m := range merged {
			if clipped.start < m.end && clipped.end > m.start {
				if clipped.start >= m.start {
					clipped.start = m.end
				}
			}
		}
		if clipped.start < clipped.end {
			merged = append(merged, clipped)
		}
	}

	// Build output by interleaving non-span and span segments
	var buf strings.Builder
	pos := 0
	for _, r := range merged {
		if pos < r.start {
			nonSpan := ln.text[pos:r.start]
			buf.WriteString(applyStyle(nonSpan, globalNonSpanCode))
		}
		spanText := ln.text[r.start:r.end]
		style := r.style
		if style == nil {
			style = globalSpanCode
		}
		buf.WriteString(applyStyle(spanText, style))
		pos = r.end
	}
	if pos < len(ln.text) {
		nonSpan := ln.text[pos:]
		buf.WriteString(applyStyle(nonSpan, globalNonSpanCode))
	}
	return buf.String()
}

func (r *Renderer) Write(w io.Writer, src []byte, labels []Label) error {
	if len(src) == 0 && len(labels) == 0 {
		return nil
	}

	for i := range labels {
		if labels[i].Marker == 0 {
			labels[i].Marker = MarkerDash
		}
		lbl := labels[i]
		if lbl.Span.Start < 0 {
			return fmt.Errorf("label %d: span start %d is negative", i, lbl.Span.Start)
		}
		if lbl.Span.End > len(src) {
			return fmt.Errorf("label %d: span end %d exceeds source length %d", i, lbl.Span.End, len(src))
		}
		if lbl.Span.Start >= lbl.Span.End {
			return fmt.Errorf("label %d: span start %d must be less than end %d", i, lbl.Span.Start, lbl.Span.End)
		}
	}

	lines := parseLines(src)
	if len(lines) == 0 {
		return nil
	}

	labelMap := mapLabelsToLines(lines, labels)

	maxLineNum := lines[len(lines)-1].number
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))
	padding := strings.Repeat(" ", lineNumWidth)

	for li, ln := range lines {
		ll := labelMap[li]
		if len(ll) > 0 {
			sort.Slice(ll, func(i, j int) bool {
				if ll[i].startInLine != ll[j].startInLine {
					return ll[i].startInLine < ll[j].startInLine
				}
				si := ll[i].label.Span
				sj := ll[j].label.Span
				return (si.End - si.Start) < (sj.End - sj.Start)
			})
		}

		lineNumStr := fmt.Sprintf("%*d", lineNumWidth, ln.number)

		// If this line has labels, use the first label's style for line number and separator
		var labelLineNumStyle, labelSepStyle StyleFunc
		if len(ll) > 0 {
			labelLineNumStyle = ll[0].label.Style.LineNumber
			labelSepStyle = ll[0].label.Style.Separator
		}

		styledLineNum := resolveStyle(lineNumStr, labelLineNumStyle, r.Style.LineNumber)
		styledSep := resolveStyle("|", labelSepStyle, r.Style.Separator)

		styledCode := segmentCodeLine(ln, ll, r.Style.SpanCode, r.Style.NonSpanCode)
		if _, err := fmt.Fprintf(w, "%s %s %s\n", styledLineNum, styledSep, styledCode); err != nil {
			return err
		}

		if len(ll) == 0 {
			continue
		}

		for _, lbl := range ll {
			expandedStart := ln.offsetMap[lbl.startInLine]
			expandedEnd := ln.offsetMap[lbl.endInLine]
			textBefore := ln.text[:expandedStart]
			textInSpan := ln.text[expandedStart:expandedEnd]
			leadingWidth := runewidth.StringWidth(textBefore)
			markerWidth := runewidth.StringWidth(textInSpan)

			leading := strings.Repeat(" ", leadingWidth)
			markers := strings.Repeat(string(lbl.label.Marker), markerWidth)

			styledMarkers := resolveStyle(markers, lbl.label.Style.Marker, r.Style.Marker)
			styledSep := resolveStyle("|", lbl.label.Style.Separator, r.Style.Separator)

			if lbl.isLastLine && lbl.label.Text != "" {
				styledText := resolveStyle(lbl.label.Text, lbl.label.Style.LabelText, r.Style.LabelText)
				if _, err := fmt.Fprintf(w, "%s %s %s%s %s\n", padding, styledSep, leading, styledMarkers, styledText); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s %s %s%s\n", padding, styledSep, leading, styledMarkers); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
