package main

import (
	"strings"
	"testing"
)

func mustProcess(t *testing.T, src string, opts Options) (string, bool) {
	t.Helper()
	return ProcessText(src, opts)
}

func TestFormatBasic(t *testing.T) {
	in := "| Name | Age |\n|---|---|\n| Alice | 30 |\n| Bob | 7 |\n"
	want := "| Name  | Age |\n" +
		"| ----- | --- |\n" +
		"| Alice | 30  |\n" +
		"| Bob   | 7   |\n"

	got, changed := mustProcess(t, in, Options{})
	if !changed {
		t.Fatal("expected changed=true")
	}
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestAlreadyFormattedIsUnchanged(t *testing.T) {
	in := "| Name  | Age |\n| ----- | --- |\n| Alice | 30  |\n"
	got, changed := mustProcess(t, in, Options{})
	if changed {
		t.Fatalf("expected no change, got:\n%s", got)
	}
	if got != in {
		t.Fatalf("got:\n%q\nwant:\n%q", got, in)
	}
}

func TestCJKWidth(t *testing.T) {
	in := "| 名字 | x |\n| --- | --- |\n| 李四 | 1 |\n"
	want := "| 名字 | x   |\n" +
		"| ---- | --- |\n" +
		"| 李四 | 1   |\n"

	got, _ := mustProcess(t, in, Options{})
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEscapedPipeStaysInCell(t *testing.T) {
	in := "| k | v |\n| --- | --- |\n| a \\| b | 2 |\n"
	got, _ := mustProcess(t, in, Options{})

	if !strings.Contains(got, `| a \| b | 2   |`) {
		t.Fatalf("escaped pipe was split or misaligned:\n%s", got)
	}
	cells := splitCells(`| a \| b | 2 |`)
	if len(cells) != 2 || cells[0] != `a \| b` || cells[1] != "2" {
		t.Fatalf("splitCells = %#v", cells)
	}
}

func TestAlignmentMarkersRoundTrip(t *testing.T) {
	in := "|h1|h2|h3|\n|:--|--:|:-:|\n|a|b|c|\n"
	want := "| h1  |  h2 | h3  |\n" +
		"| :-- | --: | :-: |\n" +
		"| a   |   b |  c  |\n"

	got, _ := mustProcess(t, in, Options{})
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSortNumeric(t *testing.T) {
	in := "| n | name |\n| --- | --- |\n| 10 | ten |\n| 2 | two |\n| 1 | one |\n"
	want := "| n   | name |\n" +
		"| --- | ---- |\n" +
		"| 1   | one  |\n" +
		"| 2   | two  |\n" +
		"| 10  | ten  |\n"

	got, _ := mustProcess(t, in, Options{SortCol: 1, Numeric: true})
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSortLexical(t *testing.T) {
	in := "| n |\n| --- |\n| 10 |\n| 2 |\n| 1 |\n"
	want := "| n   |\n| --- |\n| 1   |\n| 10  |\n| 2   |\n"

	got, _ := mustProcess(t, in, Options{SortCol: 1})
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestIdempotent(t *testing.T) {
	in := "| 名字 | 值 |\n|:--|--:|\n| 张三 | 100 |\n| bob | 7 |\n"
	once, _ := mustProcess(t, in, Options{})
	twice, changed := mustProcess(t, once, Options{})
	if changed {
		t.Fatalf("second pass changed output:\n%s\n---\n%s", once, twice)
	}
	if once != twice {
		t.Fatalf("not idempotent:\n%q\n%q", once, twice)
	}
}

func TestFencedCodeIsUntouched(t *testing.T) {
	fence := strings.Repeat("`", 3)
	in := fence + "text\n| a | b |\n|---|---|\n|1|2|\n" + fence + "\n"
	got, changed := mustProcess(t, in, Options{})
	if changed || got != in {
		t.Fatalf("fenced block was modified:\n%s", got)
	}
}

func TestTextAroundTablesSurvives(t *testing.T) {
	in := "# Title\n\n|a|b|\n|-|-|\n|1|2|\n\nTrailing paragraph.\n"
	got, _ := mustProcess(t, in, Options{})
	want := "# Title\n\n| a   | b   |\n| --- | --- |\n| 1   | 2   |\n\nTrailing paragraph.\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestTwoTablesInOneDocument(t *testing.T) {
	in := "|a|\n|-|\n|1|\n\nmiddle\n\n|bb|cc|\n|--|--:|\n|x|yy|\n"
	got, _ := mustProcess(t, in, Options{})
	if !strings.Contains(got, "| 1   |") && !strings.Contains(got, "| 1   |") {
		t.Fatalf("first table not formatted:\n%s", got)
	}
	if !strings.Contains(got, "| x   |  yy |") {
		t.Fatalf("second table not formatted:\n%s", got)
	}
}

func TestSetextHeadingIsNotATable(t *testing.T) {
	in := "Title\n---\n"
	got, changed := mustProcess(t, in, Options{})
	if changed || got != in {
		t.Fatalf("setext heading was treated as a table:\n%q", got)
	}
	if isDelimRow("---") {
		t.Fatal("isDelimRow(---) should be false without a pipe")
	}
}

func TestRaggedRowsArePadded(t *testing.T) {
	in := "| a | b | c |\n| --- | --- | --- |\n| 1 |\n| 1 | 2 | 3 | 4 |\n"
	got, _ := mustProcess(t, in, Options{})
	want := "| a   | b   | c   |\n" +
		"| --- | --- | --- |\n" +
		"| 1   |     |     |\n" +
		"| 1   | 2   | 3   |\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestLeadingIndentIsPreserved(t *testing.T) {
	in := "  | a | b |\n  | - | - |\n  | 1 | 2 |\n"
	got, _ := mustProcess(t, in, Options{})
	if !strings.HasPrefix(got, "  | ") {
		t.Fatalf("indent lost:\n%q", got)
	}
}

func TestCRLFPreserved(t *testing.T) {
	in := "# doc\r\n\r\n|a|b|\r\n|-|-|\r\n|1|2|\r\n"
	want := "# doc\r\n\r\n| a   | b   |\r\n| --- | --- |\r\n| 1   | 2   |\r\n"

	got, _ := mustProcess(t, in, Options{})
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Fatalf("output contains a lone LF: %q", got)
	}
}

func TestCRLFAlreadyFormattedIsUnchanged(t *testing.T) {
	in := "# doc\r\n\r\n| a   | b   |\r\n| --- | --- |\r\n| 1   | 2   |\r\n"
	got, changed := mustProcess(t, in, Options{})
	if changed || got != in {
		t.Fatalf("changed=%v got=%q want=%q", changed, got, in)
	}
}

func TestMixedLineEndingsAreNotNormalized(t *testing.T) {
	// Mixed endings are ambiguous; mdt must not silently rewrite the file.
	in := "# doc\n\r\nplain line\r\n"
	got, changed := mustProcess(t, in, Options{})
	if changed || got != in {
		t.Fatalf("changed=%v got=%q", changed, got)
	}
}

func TestHasLoneLF(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"a\r\nb\r\n", false},
		{"a\nb\n", true},
		{"\n", true},
		{"a\r\nb\n", true},
		{"", false},
	}
	for _, c := range cases {
		if got := hasLoneLF(c.in); got != c.want {
			t.Fatalf("hasLoneLF(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCRLFRejected(t *testing.T) {
	// CRLF input should still parse: trailing \r is trimmed per cell.
	in := "|a|b|\r\n|-|-|\r\n|1|2|\r\n"
	got, _ := mustProcess(t, in, Options{})
	if !strings.Contains(got, "| 1   | 2   |") {
		t.Fatalf("CRLF table not formatted:\n%q", got)
	}
}

func TestParseNumber(t *testing.T) {
	cases := map[string]float64{
		"1,234":  1234,
		"**42**": 42,
		"-3.5":   -3.5,
	}
	for in, want := range cases {
		got, ok := parseNumber(in)
		if !ok || got != want {
			t.Fatalf("parseNumber(%q) = %v, %v; want %v", in, got, ok, want)
		}
	}
	if _, ok := parseNumber("n/a"); ok {
		t.Fatal("parseNumber(n/a) should fail")
	}
}
