package main

import "testing"

func TestRuneWidth(t *testing.T) {
	cases := []struct {
		in   rune
		want int
	}{
		{'a', 1},
		{' ', 1},
		{'名', 2},
		{'字', 2},
		{'あ', 2},
		{'한', 2},
		{'Ａ', 2},
		{'😀', 2},
		{'\u0301', 0}, // combining acute
		{'\u200b', 0}, // zero-width space
		{'\t', 0},
	}
	for _, c := range cases {
		if got := runeWidth(c.in); got != c.want {
			t.Errorf("runeWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestDisplayWidth(t *testing.T) {
	cases := map[string]int{
		"hello":   5,
		"名字":      4,
		"名字abc":   7,
		"e\u0301": 1,
		"":        0,
		"a😀b":     4,
	}
	for in, want := range cases {
		if got := displayWidth(in); got != want {
			t.Errorf("displayWidth(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestWideRangesSorted(t *testing.T) {
	for i := 1; i < len(wideRanges); i++ {
		if wideRanges[i].lo <= wideRanges[i-1].hi {
			t.Fatalf("wideRanges not sorted at %d: %#v", i, wideRanges[i])
		}
	}
}
