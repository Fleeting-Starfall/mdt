package main

import (
	"strings"
	"testing"
)

// cellsWithin fails when any rendered cell segment is wider than width display
// columns. Each row is one line of Markdown; a wrapped cell holds several
// segments joined by "<br>".
func cellsWithin(t *testing.T, doc string, width int) {
	t.Helper()
	for _, line := range strings.Split(doc, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		for _, cell := range parts[1 : len(parts)-1] {
			for _, seg := range strings.Split(cell, "<br>") {
				if w := displayWidth(strings.TrimSpace(seg)); w > width {
					t.Fatalf("cell %q is %d columns, limit is %d\n%s", seg, w, width, doc)
				}
			}
		}
	}
}

func TestMaxWidthWrapsLongCells(t *testing.T) {
	in := "| name | description |\n| --- | --- |\n| a | one two three four five |\n"
	got, _ := ProcessText(in, Options{MaxWidth: 12})

	if !strings.Contains(got, "<br>") {
		t.Fatalf("expected a wrapped cell, got:\n%s", got)
	}
	cellsWithin(t, got, 12)
}

func TestMaxWidthKeepsShortColumnsNarrow(t *testing.T) {
	in := "| a | b |\n| --- | --- |\n| 1 | 2 |\n"
	got, _ := ProcessText(in, Options{MaxWidth: 40})
	want := "| a   | b   |\n| --- | --- |\n| 1   | 2   |\n"
	if got != want {
		t.Fatalf("short table changed:\n%q\nwant\n%q", got, want)
	}
}

func TestMaxWidthCountsDisplayWidthForCJK(t *testing.T) {
	in := "| k | v |\n| --- | --- |\n| a | 一二三四五六 |\n"
	got, _ := ProcessText(in, Options{MaxWidth: 6})

	if !strings.Contains(got, "<br>") {
		t.Fatalf("CJK cell was not wrapped:\n%s", got)
	}
	cellsWithin(t, got, 6)
}

func TestMaxWidthIsIdempotent(t *testing.T) {
	in := "| name | description |\n| --- | --- |\n| a | one two three four five six seven |\n"
	once, _ := ProcessText(in, Options{MaxWidth: 10})
	twice, changed := ProcessText(once, Options{MaxWidth: 10})
	if changed || twice != once {
		t.Fatalf("not idempotent\nonce:\n%s\ntwice:\n%s", once, twice)
	}
}

func TestMaxWidthRespectsAlignment(t *testing.T) {
	in := "| l | r | c |\n| :-- | --: | :-: |\n| aaa bbb ccc | ddd eee fff | ggg hhh iii |\n"
	got, _ := ProcessText(in, Options{MaxWidth: 6})

	delim := strings.Split(strings.Split(got, "\n")[1], "|")
	left, right, center := strings.TrimSpace(delim[1]), strings.TrimSpace(delim[2]), strings.TrimSpace(delim[3])
	if !strings.HasPrefix(left, ":") {
		t.Fatalf("left column lost its colon: %q", left)
	}
	if !strings.HasSuffix(right, ":") {
		t.Fatalf("right column lost its colon: %q", right)
	}
	if !strings.HasPrefix(center, ":") || !strings.HasSuffix(center, ":") {
		t.Fatalf("center column lost a colon: %q", center)
	}
	cellsWithin(t, got, 6)
}

func TestWrapLinePrefersSpaces(t *testing.T) {
	got := wrapLine("alpha beta gamma", 9)
	want := []string{"alpha", "beta", "gamma"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("wrapLine = %q, want %q", got, want)
	}
}

func TestWrapLineHardBreaksLongWord(t *testing.T) {
	got := wrapLine("abcdefghij", 4)
	for _, piece := range got {
		if displayWidth(piece) > 4 {
			t.Fatalf("piece too wide: %q", piece)
		}
	}
	if strings.Join(got, "") != "abcdefghij" {
		t.Fatalf("characters lost: %q", got)
	}
}

func TestWrapLineKeepsEscapedPipeWhole(t *testing.T) {
	got := wrapLine(`aaaa \| bbbb`, 6)
	for _, piece := range got {
		if strings.HasSuffix(piece, `\`) {
			t.Fatalf("escape split from its pipe: %q", got)
		}
	}
}

func TestMaxWidthBelowThreeIsClamped(t *testing.T) {
	in := "| abcdef |\n| --- |\n| ghijkl |\n"
	got, _ := ProcessText(in, Options{MaxWidth: 1})
	if !strings.Contains(got, "| --- |") {
		t.Fatalf("delimiter row below minimum:\n%s", got)
	}
}
