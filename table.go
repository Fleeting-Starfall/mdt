package main

import (
	"sort"
	"strconv"
	"strings"
)

// Align is the alignment declared by a table's delimiter row.
type Align int

const (
	AlignNone Align = iota
	AlignLeft
	AlignRight
	AlignCenter
)

// Options controls how tables are formatted.
type Options struct {
	// SortCol is a 1-based column index used to reorder body rows.
	// Zero keeps the original row order.
	SortCol int
	// Numeric compares the sort column as numbers instead of strings.
	Numeric bool
}

// Table is a parsed Markdown table: a header row, the alignment row, and body rows.
type Table struct {
	Header []string
	Align  []Align
	Rows   [][]string
	// Indent is the whitespace that preceded the header row.
	Indent string
}

// ProcessText reformats every table in src and reports whether anything changed.
// Content inside fenced code blocks is never touched.
func ProcessText(src string, opts Options) (string, bool) {
	// A CRLF document must stay CRLF: rewriting the whole file to LF turns a
	// one-table edit into a whole-file diff. Mixed endings are left as they are.
	crlf := strings.Contains(src, "\r\n") && !hasLoneLF(src)
	if crlf {
		src = strings.ReplaceAll(src, "\r\n", "\n")
	}

	trailingNewline := strings.HasSuffix(src, "\n")
	lines := strings.Split(src, "\n")
	if trailingNewline && len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	out := make([]string, 0, len(lines))
	changed := false
	var fence byte // 0 when not inside a fenced code block

	for i := 0; i < len(lines); {
		line := lines[i]

		if fence != 0 {
			if isFenceClose(line, fence) {
				fence = 0
			}
			out = append(out, line)
			i++
			continue
		}
		if c, ok := fenceOpen(line); ok {
			fence = c
			out = append(out, line)
			i++
			continue
		}

		if i+1 < len(lines) && hasPipe(line) && isDelimRow(lines[i+1]) {
			end := i + 2
			for end < len(lines) && hasPipe(lines[end]) {
				if _, ok := fenceOpen(lines[end]); ok {
					break
				}
				end++
			}
			block := lines[i:end]
			formatted := parseBlock(block).Format(opts)
			if formatted != strings.Join(block, "\n") {
				changed = true
			}
			out = append(out, strings.Split(formatted, "\n")...)
			i = end
			continue
		}

		out = append(out, line)
		i++
	}

	res := strings.Join(out, "\n")
	if trailingNewline {
		res += "\n"
	}
	if crlf {
		res = strings.ReplaceAll(res, "\n", "\r\n")
	}
	return res, changed
}

// hasLoneLF reports whether s contains a newline that is not part of a CRLF
// pair, which means the document is not uniformly CRLF.
func hasLoneLF(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			return true
		}
	}
	return false
}

// parseBlock builds a Table from a header line, a delimiter line and body lines.
func parseBlock(lines []string) *Table {
	t := &Table{}

	header := lines[0]
	t.Indent = header[:len(header)-len(strings.TrimLeft(header, " \t"))]

	t.Header = splitCells(header)
	t.Align = parseDelims(lines[1])
	if len(t.Align) < len(t.Header) {
		for len(t.Align) < len(t.Header) {
			t.Align = append(t.Align, AlignNone)
		}
	} else if len(t.Align) > len(t.Header) {
		// A malformed delimiter row may declare extra columns; drop them.
		t.Align = t.Align[:len(t.Header)]
	}

	for _, line := range lines[2:] {
		row := splitCells(line)
		if len(row) > len(t.Header) {
			row = row[:len(t.Header)]
		}
		for len(row) < len(t.Header) {
			row = append(row, "")
		}
		t.Rows = append(t.Rows, row)
	}
	return t
}

// Format renders the table with aligned columns.
func (t *Table) Format(opts Options) string {
	n := len(t.Header)
	if n == 0 {
		return ""
	}

	rows := t.Rows
	if opts.SortCol > 0 && opts.SortCol <= n {
		rows = sortRows(rows, opts.SortCol-1, opts.Numeric)
	}

	widths := make([]int, n)
	for i := range widths {
		widths[i] = 3 // a delimiter needs at least "---"
	}
	for i, h := range t.Header {
		if w := displayWidth(visible(h)); w > widths[i] {
			widths[i] = w
		}
	}
	for _, row := range rows {
		for i := 0; i < n; i++ {
			if w := displayWidth(visible(row[i])); w > widths[i] {
				widths[i] = w
			}
		}
	}

	var b strings.Builder
	writeLine := func(cells []string) {
		b.WriteString(t.Indent)
		b.WriteString("|")
		for i := 0; i < n; i++ {
			b.WriteString(" ")
			b.WriteString(pad(cells[i], widths[i], t.Align[i]))
			b.WriteString(" |")
		}
		b.WriteString("\n")
	}

	writeLine(t.Header)

	b.WriteString(t.Indent)
	b.WriteString("|")
	for i := 0; i < n; i++ {
		b.WriteString(" ")
		b.WriteString(delimCell(widths[i], t.Align[i]))
		b.WriteString(" |")
	}
	b.WriteString("\n")

	for _, row := range rows {
		writeLine(row)
	}

	return strings.TrimSuffix(b.String(), "\n")
}

// delimCell renders one cell of the delimiter row, e.g. ":---" or ":-:".
func delimCell(width int, a Align) string {
	if width < 3 {
		width = 3
	}
	switch a {
	case AlignLeft:
		return ":" + strings.Repeat("-", width-1)
	case AlignRight:
		return strings.Repeat("-", width-1) + ":"
	case AlignCenter:
		return ":" + strings.Repeat("-", width-2) + ":"
	default:
		return strings.Repeat("-", width)
	}
}

// pad adds spaces to cell so its visible width becomes width.
func pad(cell string, width int, a Align) string {
	gap := width - displayWidth(visible(cell))
	if gap <= 0 {
		return cell
	}
	switch a {
	case AlignRight:
		return strings.Repeat(" ", gap) + cell
	case AlignCenter:
		left := gap / 2
		return strings.Repeat(" ", left) + cell + strings.Repeat(" ", gap-left)
	default:
		return cell + strings.Repeat(" ", gap)
	}
}

// visible returns the text a reader sees inside a cell, so that escape
// sequences are not counted as extra characters.
func visible(cell string) string {
	return strings.ReplaceAll(cell, `\|`, "|")
}

// splitCells splits a table row into trimmed cells on unescaped pipes.
// Outer pipes are optional; "\|" is treated as a literal pipe.
func splitCells(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	runes := []rune(line)
	var cells []string
	var cur strings.Builder
	firstWasPipe := runes[0] == '|'
	lastWasPipe := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\\' && i+1 < len(runes) && runes[i+1] == '|' {
			cur.WriteRune('\\')
			cur.WriteRune('|')
			i++
			continue
		}
		if r == '|' {
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
			lastWasPipe = i == len(runes)-1
			continue
		}
		cur.WriteRune(r)
	}
	cells = append(cells, strings.TrimSpace(cur.String()))

	if firstWasPipe && len(cells) > 0 {
		cells = cells[1:]
	}
	if lastWasPipe && len(cells) > 0 {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// parseDelims extracts alignment from a delimiter row such as "|:---|---:|".
func parseDelims(line string) []Align {
	cells := splitCells(line)
	aligns := make([]Align, 0, len(cells))
	for _, c := range cells {
		c = strings.TrimSpace(c)
		left := strings.HasPrefix(c, ":")
		right := strings.HasSuffix(c, ":")
		switch {
		case left && right:
			aligns = append(aligns, AlignCenter)
		case left:
			aligns = append(aligns, AlignLeft)
		case right:
			aligns = append(aligns, AlignRight)
		default:
			aligns = append(aligns, AlignNone)
		}
	}
	return aligns
}

// isDelimRow reports whether line is a table delimiter row.
func isDelimRow(line string) bool {
	if !hasPipe(line) {
		return false
	}
	cells := splitCells(line)
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c == "" {
			return false
		}
		body := strings.TrimSuffix(strings.TrimPrefix(c, ":"), ":")
		if len(body) < 1 {
			return false
		}
		for i := 0; i < len(body); i++ {
			if body[i] != '-' {
				return false
			}
		}
	}
	return true
}

// hasPipe reports whether line contains a pipe that is not escaped.
func hasPipe(line string) bool {
	for i := 0; i < len(line); i++ {
		if line[i] != '|' {
			continue
		}
		n := 0
		for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
			n++
		}
		if n%2 == 0 {
			return true
		}
	}
	return false
}

// fenceOpen reports whether line opens a fenced code block and returns its marker.
func fenceOpen(line string) (byte, bool) {
	s := strings.TrimLeft(line, " \t")
	for _, c := range []byte{'`', '~'} {
		if len(s) >= 3 && s[0] == c && s[1] == c && s[2] == c {
			return c, true
		}
	}
	return 0, false
}

// isFenceClose reports whether line closes a fence opened with marker c.
func isFenceClose(line string, c byte) bool {
	s := strings.TrimSpace(line)
	if len(s) < 3 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != c {
			return false
		}
	}
	return true
}

// sortRows returns a stable copy of rows ordered by column col.
func sortRows(rows [][]string, col int, numeric bool) [][]string {
	out := make([][]string, len(rows))
	copy(out, rows)
	if numeric {
		sort.SliceStable(out, func(i, j int) bool {
			a, aok := parseNumber(out[i][col])
			b, bok := parseNumber(out[j][col])
			if aok && bok {
				return a < b
			}
			return strings.TrimSpace(out[i][col]) < strings.TrimSpace(out[j][col])
		})
		return out
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.TrimSpace(out[i][col]) < strings.TrimSpace(out[j][col])
	})
	return out
}

// parseNumber reads a number out of a cell, ignoring Markdown emphasis and
// thousands separators.
func parseNumber(cell string) (float64, bool) {
	s := strings.TrimSpace(cell)
	s = strings.Trim(s, "`*_ ")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "_", "")
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
