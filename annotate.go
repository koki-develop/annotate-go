// Package annotate renders annotated source code with line numbers and labeled markers.
//
// It is designed for displaying diagnostic messages (errors, warnings, hints) that
// point to specific byte ranges within source code, similar to the output of compilers
// like rustc or Go's go vet.
//
// Basic usage:
//
//	src := []byte(`name = "Alice"
//	age = 30
//	`)
//	labels := []annotate.Label{
//		{Span: annotate.Span{Start: 0, End: 14}, Marker: annotate.MarkerDash, Text: "string field"},
//		{Span: annotate.Span{Start: 15, End: 23}, Marker: annotate.MarkerTilde, Text: "integer field"},
//	}
//	r := annotate.New()
//	output, _ := r.Render(src, labels)
//	fmt.Print(output)
//
// Output:
//
//	1 | name = "Alice"
//	  | -------------- string field
//	2 | age = 30
//	  | ~~~~~~~~ integer field
package annotate

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Span represents a half-open byte range [Start, End) within the source code.
// Both Start and End are zero-based byte offsets. Start must be less than End.
type Span struct {
	Start int
	End   int
}

// LabelMarker is the character used to underline the annotated span.
// Any rune can be used as a marker, but the package provides common presets.
type LabelMarker rune

const (
	// MarkerDash uses '-' as the marker character.
	MarkerDash LabelMarker = '-'
	// MarkerTilde uses '~' as the marker character.
	MarkerTilde LabelMarker = '~'
	// MarkerCaret uses '^' as the marker character.
	MarkerCaret LabelMarker = '^'
)

// Label defines an annotation for a byte range within the source code.
// Each label points to a [Span], uses a [LabelMarker] to underline that span,
// and optionally displays descriptive text after the markers.
//
// If Marker is zero, [MarkerDash] is used by default.
// If Style is set, its non-nil fields override the global [Style] for this label.
// Labels may span multiple lines; the text is displayed on the last line of the span.
type Label struct {
	Span   Span
	Marker LabelMarker
	Text   string
	Style  LabelStyle
	// Before overrides [Renderer.Before] for this label. Nil falls back to the Renderer default.
	Before *int
	// After overrides [Renderer.After] for this label. Nil falls back to the Renderer default.
	After *int
}

// Renderer renders annotated source code. Use [New] to create a Renderer.
type Renderer struct {
	// Style controls the visual styling applied to each element of the output.
	// If zero, no styling is applied.
	Style Style

	// Before is the default number of lines to display before each labeled line.
	Before int
	// After is the default number of lines to display after each labeled line.
	After int
}

// New creates a new [Renderer] with the given options.
func New(opts ...Option) *Renderer {
	r := &Renderer{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Render renders the annotated source code and returns it as a string.
//
// Each label's [Span] must satisfy: 0 <= Start < End <= len(src).
// An error is returned if any label has an invalid span.
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

type visibleLines struct {
	flags  []bool
	first  int // first visible line index (-1 if none)
	last   int // last visible line index (-1 if none)
	maxNum int // maximum visible line number
}

func computeVisibleLines(lines []line, labelMap map[int][]lineLabel, defaultBefore, defaultAfter int) visibleLines {
	flags := make([]bool, len(lines))
	for li, lls := range labelMap {
		for _, ll := range lls {
			before := defaultBefore
			after := defaultAfter
			if ll.label.Before != nil {
				before = max(0, *ll.label.Before)
			}
			if ll.label.After != nil {
				after = max(0, *ll.label.After)
			}
			low := max(0, li-before)
			high := min(len(lines)-1, li+after)
			for j := low; j <= high; j++ {
				flags[j] = true
			}
		}
	}

	first, last := -1, -1
	maxNum := 0
	for li, ln := range lines {
		if flags[li] {
			if ln.number > maxNum {
				maxNum = ln.number
			}
			if first < 0 {
				first = li
			}
			last = li
		}
	}
	return visibleLines{flags: flags, first: first, last: last, maxNum: maxNum}
}

// Write renders the annotated source code and writes the output to w.
//
// Each label's [Span] must satisfy: 0 <= Start < End <= len(src).
// An error is returned if any label has an invalid span or if writing to w fails.
// Only lines covered by labels (plus surrounding Before/After context lines) are output.
// If labels is empty, nothing is written.
func validateLabels(src []byte, labels []Label) error {
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
	return nil
}

func (r *Renderer) Write(w io.Writer, src []byte, labels []Label) error {
	if len(labels) == 0 {
		return nil
	}

	if err := validateLabels(src, labels); err != nil {
		return err
	}

	lines := parseLines(src)
	if len(lines) == 0 {
		return nil
	}

	labelMap := mapLabelsToLines(lines, labels)
	vis := computeVisibleLines(lines, labelMap, r.Before, r.After)

	return r.writeOutput(w, lines, labelMap, vis)
}

func (r *Renderer) writeOutput(w io.Writer, lines []line, labelMap map[int][]lineLabel, vis visibleLines) error {
	if vis.maxNum == 0 {
		return nil
	}
	lineNumWidth := len(fmt.Sprintf("%d", vis.maxNum))
	padding := strings.Repeat(" ", lineNumWidth)

	writeEllipsis := func() error {
		ellipsis := "..."
		if lineNumWidth > 3 {
			ellipsis += strings.Repeat(" ", lineNumWidth-3)
		}
		styledEllipsis := applyStyle(ellipsis, r.Style.Ellipsis)
		_, err := fmt.Fprintf(w, "%s\n", styledEllipsis)
		return err
	}

	if vis.first > 0 {
		if err := writeEllipsis(); err != nil {
			return err
		}
	}

	prevVisibleIdx := -1
	for li, ln := range lines {
		if !vis.flags[li] {
			continue
		}

		if prevVisibleIdx >= 0 && li > prevVisibleIdx+1 {
			if err := writeEllipsis(); err != nil {
				return err
			}
		}
		prevVisibleIdx = li

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

		if len(ll) > 0 {
			if err := r.writeMarkerLines(w, padding, ln, ll); err != nil {
				return err
			}
		}
	}

	if vis.last >= 0 && vis.last < len(lines)-1 {
		if err := writeEllipsis(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Renderer) writeMarkerLines(w io.Writer, padding string, ln line, ll []lineLabel) error {
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
	return nil
}
