# mdt

`mdt` aligns Markdown tables from the command line: it re-pads every cell so
columns line up, normalizes the `---|:---:|` delimiter row, and leaves the rest
of the document alone.

It counts **display width** instead of bytes, so tables mixing Latin and CJK text
line up in terminals and editors that render `名` as two columns wide.

```
$ mdt notes.md
| name     | value | note |
| :------- | :---: | ---: |
| alpha    |   1   | fast |
| 中文名字 |  100  | 很慢 |
| b        |  22   |    x |
```

## Install

```sh
go install github.com/Fleeting-Starfall/mdt@latest
```

Or build from a checkout:

```sh
git clone https://github.com/Fleeting-Starfall/mdt
cd mdt && go build .
```

No dependencies beyond the Go standard library.

## Usage

```
mdt [flags] [file ...]
```

With no files it reads standard input. Output goes to standard output unless
`-w` is given.

| flag | meaning                                          |
| :--- | :----------------------------------------------- |
| `-w` | rewrite each file in place                       |
| `-c` | check mode: exit 1 if any file would change      |
| `-s` | sort body rows by the given 1-based column       |
| `-n` | numeric instead of string comparison (with `-s`) |
| `-q` | quiet: no per-file messages in check mode        |
| `-V` | print the version                                |

### Examples

Format every Markdown file in a repo:

```sh
mdt -w README.md docs/*.md
```

Fail CI when a table is not formatted:

```sh
mdt -c README.md docs/*.md
```

Sort a scoreboard by points, highest first — pipe through the shell for that:

```sh
mdt -s 3 -n leaderboard.md | sort -t'|' -k4 -nr
```

## What it does, precisely

- A table is a line containing an unescaped `|` followed by a delimiter row
  (`| --- | :---: | ---: |`).
- Cells are split on unescaped pipes; `\|` stays inside the cell.
- Outer pipes are optional, ragged rows are padded (or truncated) to the header
  width, and the leading indentation of the header line is preserved.
- Delimiter rows are re-emitted at the exact column width, keeping the `:`
  alignment markers (`---`, `:--`, `--:`, `:-:`).
- Tables inside fenced code blocks (```` ``` ```` or `~~~`) are skipped, so
  documentation that shows Markdown examples is safe.
- Everything that is not a table is copied through byte for byte.

## Limitations

- Single-column tables written without any outer pipe are not detected, because
  they are indistinguishable from plain text.
- Pipe characters inside inline code spans must still be escaped as `\|`
  (that is what CommonMark requires anyway).
- East Asian width is approximated with a compact range table rather than the
  full Unicode `EastAsianWidth.txt`. Ambiguous-width characters count as one
  column.

## Development

```sh
go test ./...
go vet ./...
gofmt -l .
```

## License

MIT — see [LICENSE](LICENSE).
