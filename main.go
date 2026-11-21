// Package main implements mdt, a formatter for GitHub-flavored Markdown tables.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("mdt", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		write   = fs.Bool("w", false, "rewrite each file in place")
		check   = fs.Bool("c", false, "check only: exit 1 if a file would change")
		sortCol = fs.Int("s", 0, "sort body rows by 1-based column N")
		numeric = fs.Bool("n", false, "compare numerically when sorting")
		quiet   = fs.Bool("q", false, "suppress output in check mode")
		showVer = fs.Bool("V", false, "print version and exit")
	)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `mdt %s - align Markdown tables

usage: mdt [flags] [file ...]

Reads Markdown from files (or stdin), reformats every GFM table it finds,
and writes the result to stdout. Text outside tables is left untouched.

flags:
`, version)
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), `
examples:
  mdt README.md                 # print formatted result
  mdt -w README.md docs/*.md    # format files in place
  mdt -c README.md              # exit 1 if README.md is not formatted
  mdt -s 2 -n data.md           # sort rows by column 2, numerically

exit codes:
  0  success (or, with -c, everything is already formatted)
  1  -c found a file that needs formatting
  2  usage or I/O error
`)
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *showVer {
		fmt.Println("mdt " + version)
		return 0
	}
	if *write && fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "mdt: -w requires at least one file")
		return 2
	}
	if *write && *check {
		fmt.Fprintln(os.Stderr, "mdt: -w and -c are mutually exclusive")
		return 2
	}

	opts := Options{SortCol: *sortCol, Numeric: *numeric}
	if opts.SortCol < 0 {
		fmt.Fprintln(os.Stderr, "mdt: -s must be a positive column number")
		return 2
	}

	files := fs.Args()
	if len(files) == 0 {
		src, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mdt: %v\n", err)
			return 2
		}
		out, _ := ProcessText(string(src), opts)
		fmt.Print(out)
		return 0
	}

	needsFormat := false
	for _, name := range files {
		src, err := os.ReadFile(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mdt: %v\n", err)
			return 2
		}
		out, changed := ProcessText(string(src), opts)
		switch {
		case *check:
			if changed {
				needsFormat = true
				if !*quiet {
					fmt.Fprintf(os.Stderr, "mdt: would reformat %s\n", name)
				}
			}
		case *write:
			if changed {
				info, statErr := os.Stat(name)
				mode := os.FileMode(0o644)
				if statErr == nil {
					mode = info.Mode().Perm()
				}
				if err := os.WriteFile(name, []byte(out), mode); err != nil {
					fmt.Fprintf(os.Stderr, "mdt: %v\n", err)
					return 2
				}
				if !*quiet {
					fmt.Fprintf(os.Stderr, "mdt: formatted %s\n", name)
				}
			}
		default:
			fmt.Print(out)
		}
	}
	if needsFormat {
		return 1
	}
	return 0
}
