package main

// #cgo LDFLAGS: -lreadline
// #include <stdlib.h>
// #include <readline/readline.h>
//
// void
// init_rl()
// {
//     /* use tab for indentation, not for autocompletion */
//     rl_bind_key('\t', rl_insert);
// }
//
import "C"
import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"unicode"
	"unsafe"

	"github.com/kr/text"
)

var (
	length     = flag.Int("l", 120, "maximum length of an output line")
	tabstop    = flag.Int("t", 4, "number of spaces of a tab")
	join       = flag.Bool("j", false, "join short lines when wrapping text")
	appendFile = flag.Bool("a", false, "append to file instead of overwriting")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: ted [-l N] [-t N] [-j] [-a] [file]

Ted is a line-oriented text editor.

It reads each input line using readline(3) and its text editing facilities.
The text, is then written to file, filling and indenting lines like fmt(1).

Long lines are folded to fit the maximum line length. Short lines are not joined
unless the previous line ends with a slash \ or flag -j is set.

Initial indentation of lines is preserved. Lines that are indented only with tabs
are formatted with margins both at the left and right ends.

Lines that contain tabular data, i.e data separated with tabs are formatted
using elastic tabstops http://nickgravgaard.com/elastictabstops/index.html.

Ted writes the output to file, if specified, otherwise to stdout. Currently ted
does not support editing of existing files and by default it overwrites the file.
Use -a if you want to append output to an existing file.

Flags:
`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("ted: ")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() > 1 {
		usage()
	}

	C.init_rl()

	var buf bytes.Buffer
	format(readlines(), &buf)

	w := os.Stdout
	if flag.NArg() == 1 {
		perms := os.O_WRONLY | os.O_CREATE
		if *appendFile {
			perms |= os.O_APPEND
		} else {
			perms |= os.O_TRUNC
		}

		if fout, err := os.OpenFile(flag.Arg(0), perms, 0666); err == nil {
			w = fout
			defer fout.Close()
		} else {
			log.Fatal(err)
		}
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Fatal(err)
	}
}

type line struct {
	text       string // text of the line
	indent     int    // number of spaces at the beginning of line
	indented   bool   // indent > 0
	incomplete bool   // line ended with \ (stripped from line.text)
	blank      bool   // line is empty or contains only white space
	tabular    bool   // has at least 2 columns separated by tabs
	quoted     bool   // is indented only with tabs
}

func (l *line) concat(r *line) {
	var builder strings.Builder
	builder.Grow(len(l.text) + 1 + len(r.text))
	builder.WriteString(l.text)
	builder.WriteRune(' ')
	builder.WriteString(r.text)
	l.text = builder.String()
	l.incomplete = r.incomplete
	l.blank = l.blank && r.blank

	becomesTabular := l.tabular || r.tabular || r.quoted
	l.tabular = becomesTabular
	l.quoted = l.quoted && !becomesTabular
}

// readline reads a line using readline(3). Returns the line and true on EOF
func readline() (*line, bool) {
	cstr := C.readline(nil)
	defer C.free(unsafe.Pointer(cstr))

	if cstr == nil {
		return nil, true
	}

	text := C.GoString(cstr)
	incomplete := strings.HasSuffix(text, "\\")
	if incomplete {
		text = text[0 : len(text)-1] // strip final slash
	}

	indent, indentChars, lastTab, tabCount := 0, 0, 0, 0
	inIndent := true
	for i, r := range text {
		if r == '\t' {
			tabCount++
			lastTab = i
			if inIndent {
				indent += *tabstop
				indent -= indent % *tabstop
				indentChars++
			}
		} else if unicode.IsSpace(r) {
			if inIndent {
				indent++
				indentChars++
			}
		} else {
			inIndent = false
		}
	}

	blank := inIndent
	quoted := !blank && tabCount > 0 && lastTab+1 == tabCount // all tabs at start of line
	tabular := !blank && tabCount > 0 && !quoted

	return &line{
		text:       text[indentChars:], // strip indentation
		indent:     indent,
		indented:   indent > 0,
		incomplete: incomplete,
		blank:      blank,
		tabular:    tabular,
		quoted:     quoted,
	}, false
}

// readlines reads all the input and concatenates lines where needed
func readlines() []*line {
	lines := make([]*line, 0, 32)

	var prevLine *line
	for currLine, eof := readline(); !eof; currLine, eof = readline() {
		if prevLine != nil && !currLine.blank && (prevLine.incomplete || *join) {
			prevLine.concat(currLine)
		} else {
			lines = append(lines, currLine)
			prevLine = currLine
		}
	}

	return lines
}

var (
	spaces = strings.Repeat(" ", 256)
	ats    = strings.Repeat("@", 256)
)

// format fmts all the inputs lines and outputs to the buffer
func format(lines []*line, buf *bytes.Buffer) {
	tabw := tabwriter.NewWriter(buf, *tabstop, *tabstop, 1, ' ', 0)

	for _, line := range lines {
		if line.tabular {
			tabw.Write([]byte(line.text + "\n"))
		} else {
			tabw.Flush()

			switch {
			case line.blank:
				// ignore
			case line.quoted:
				t := text.Indent(text.Wrap(line.text, *length-*tabstop*2), spaces[0:*tabstop])
				buf.WriteString(t)
			case line.indented:
				// add a prefix of @s as placeholder for the indentation, wrap and then remove it
				t := text.Wrap(ats[0:line.indent]+line.text, *length)
				buf.WriteString(spaces[0:line.indent])
				buf.WriteString(t[line.indent:])
			default:
				t := text.Wrap(line.text, *length)
				buf.WriteString(t)
			}
			buf.WriteRune('\n')
		}
	}

	tabw.Flush()
}
