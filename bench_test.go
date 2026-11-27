package main

import (
	"fmt"
	"strings"
	"testing"
)

// buildDoc returns a Markdown document with one table of rows body rows and a
// paragraph before and after it, which is the shape mdt normally sees.
func buildDoc(rows int) string {
	var b strings.Builder
	b.WriteString("# benchmark\n\nIntro paragraph that the formatter has to skip over.\n\n")
	b.WriteString("| id | name | value | note | flag |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for i := 0; i < rows; i++ {
		fmt.Fprintf(&b, "| %d | item-%d | %d | 中文备注 | ok |\n", i, i, i*7)
	}
	b.WriteString("\nTrailing paragraph.\n")
	return b.String()
}

// sink keeps the result reachable so the formatter cannot be optimised away.
var sink string

func benchProcess(b *testing.B, rows int) {
	src := buildDoc(rows)
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, _ := ProcessText(src, Options{})
		sink = out
	}
}

func BenchmarkProcess10Rows(b *testing.B)  { benchProcess(b, 10) }
func BenchmarkProcess1kRows(b *testing.B)  { benchProcess(b, 1000) }
func BenchmarkProcess10kRows(b *testing.B) { benchProcess(b, 10000) }
func BenchmarkProcess50kRows(b *testing.B) { benchProcess(b, 50000) }
