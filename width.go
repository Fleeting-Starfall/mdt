package main

import "unicode"

// wideRanges lists rune ranges whose East Asian Width is Wide (W) or
// Fullwidth (F). The table is deliberately compact: it covers the scripts and
// symbol blocks that actually show up in documentation tables, which is all a
// terminal-width estimate needs.
var wideRanges = []struct{ lo, hi rune }{
	{0x1100, 0x115F},   // Hangul Jamo (leading)
	{0x2E80, 0x303E},   // CJK radicals, Kangxi, CJK symbols
	{0x3041, 0x33FF},   // Hiragana, Katakana, Bopomofo, Hangul compat, CJK compat
	{0x3400, 0x4DBF},   // CJK unified ideographs extension A
	{0x4E00, 0x9FFF},   // CJK unified ideographs
	{0xA000, 0xA4CF},   // Yi syllables
	{0xA960, 0xA97F},   // Hangul Jamo extension A
	{0xAC00, 0xD7A3},   // Hangul syllables
	{0xF900, 0xFAFF},   // CJK compatibility ideographs
	{0xFE10, 0xFE19},   // vertical forms
	{0xFE30, 0xFE6F},   // CJK compatibility forms
	{0xFF00, 0xFF60},   // fullwidth forms
	{0xFFE0, 0xFFE6},   // fullwidth signs
	{0x16FE0, 0x16FE4}, // ideographic symbols and punctuation
	{0x17000, 0x18AFF}, // Tangut
	{0x1B000, 0x1B2FF}, // Kana supplements
	{0x1F300, 0x1F64F}, // pictographs, emoticons
	{0x1F900, 0x1F9FF}, // supplemental symbols and pictographs
	{0x20000, 0x2FFFD}, // CJK extensions B-F
	{0x30000, 0x3FFFD}, // CJK extension G
}

// displayWidth returns the number of terminal columns s occupies, counting wide
// (CJK) characters as two columns and combining marks as zero.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 32 || (r >= 0x7F && r < 0xA0):
		return 0
	case isZeroWidth(r):
		return 0
	case isWide(r):
		return 2
	default:
		return 1
	}
}

func isZeroWidth(r rune) bool {
	switch r {
	case 0x200B, 0x200C, 0x200D, 0xFEFF: // zero-width space/joiners, BOM
		return true
	}
	if r >= 0xFE00 && r <= 0xFE0F { // variation selectors
		return true
	}
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r)
}

func isWide(r rune) bool {
	lo, hi := 0, len(wideRanges)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		switch {
		case r < wideRanges[mid].lo:
			hi = mid - 1
		case r > wideRanges[mid].hi:
			lo = mid + 1
		default:
			return true
		}
	}
	return false
}
