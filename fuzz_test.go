package main

import "testing"

// FuzzProcessText checks the two properties the formatter has to hold for
// arbitrary input: it must never panic, and formatting must be idempotent
// (a second pass over its own output changes nothing).
func FuzzProcessText(f *testing.F) {
	seeds := []string{
		"",
		"\n",
		"|a|\n|-|\n|1|\n",
		"| a | b |\n| :-: | --: |\n| 1 | 2 |\n",
		"|a|b|\r\n|-|-|\r\n|1|2|\r\n",
		"```\n|a|\n|-|\n```\n",
		"|a \\| b|c|\n|---|---|\n|1|2|\n",
		"  | a |\n  | - |\n| a | b | c |\n| - |\n",
		"|a|\n|---|\n||\n",
		"正文\n\n| 名字 | 值 |\n|:--|--:|\n| 张三 | 100 |\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, src string) {
		for _, maxWidth := range []int{0, 3, 6, 12, 40} {
			out, _ := ProcessText(src, Options{MaxWidth: maxWidth})
			again, changed := ProcessText(out, Options{MaxWidth: maxWidth})
			if changed || again != out {
				t.Fatalf("not idempotent at -m %d\nin:     %q\nfirst:  %q\nsecond: %q",
					maxWidth, src, out, again)
			}
		}
	})
}
