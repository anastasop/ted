# Ted

Ted is a line oriented text editor and text formatter based on GNU readline, designed to be used
from the console.

Many times, we run some programs and we want to write a small text file based
on their output. If we use vim or emacs, they go to full screen and we no longer can see the output
to copy paste in the editor. Of course there are alternatives, like run the programs from within
the editor or open the editor in another window. But these conveniences are not always available.
What if you run something on the production server or in another machine?

Another case is  rudimentary text formatting for small files, like folding long
lines or arranging text in columns. Surely cat(1) or ed(1) cannot be used for these.

The inspiration comes from the Plan9 window system [rio](http://9p.io/magic/man2html/1/rio)
where each console window, after pressing ESC, entered hold mode and you could edit everywhere
much like turning each window into a notepad window.

Doing the full work with curses to implement the exact behavior as in rio, is too much work
and probably not worth it. Ted is a compromise. Each line is edited with readline(3)
and the final text is formatted and written to a file or to standard output. Also ted cannot be
used to edit an existing file. If you already have a file, then ed(1) or vim are better choices.
Ted is for creating new files or appending to existing ones, from the console, using some neat
text formatting conviniences.

Enjoy!

## Installation

Ted is written in Go but uses [cgo](https://golang.org/cmd/cgo/) to link with GNU readline.
You need GCC and the readline development package to build it. If you don't already have them
install them with `sudo apt-get install build-essential libreadline-dev` (debian and ubuntu),
and then install ted with `go get -u github.com/anastasop/ted`
The binary will be at $GOPATH/bin

## Usage

Checkout this demo of ted and compare the input with the output. Also `ted -h` gives a quick
overview.

Input

```
winwasher@penguin:~/src/ted$ ./ted -l 40 usage.txt
Ted is line oriented text editor.
It is powered by readline: C-a C-e etc

It is also a text formatter. Long lines like this are folded.

Short lines are not joined.

You can use a slash at the end\
to join many short lines.\

Indentation is preserved.
  - one
    - two

Lines indented with tabs are quotes:
        This is a quote, a rather long quote, that spans many lines

Finally you can use tabs for tabular output:
one     two     three
1111111111      22222222222     33333333
AAAAAAAAAAAAAAAAA       BBBBBBBBBBBBBB  CCCCCCCC

```

Output

```
winwasher@penguin:~/src/ted$ cat usage.txt
Ted is line oriented text editor.
It is powered by readline: C-a C-e etc

It is also a text formatter. Long lines
like this are folded.

Short lines are not joined.

You can use a slash at the end to join
many short lines.

Indentation is preserved.
  - one
    - two

Lines indented with tabs are quotes:
    This is a quote, a rather long
    quote, that spans many lines

Finally you can use tabs for tabular
output:
one               two            three
1111111111        22222222222    33333333
AAAAAAAAAAAAAAAAA BBBBBBBBBBBBBB CCCCCCCC

```

## License

Ted is released under the GNU public license version 3.

## Bugs

- Ted gets confused if tabstop and length are small and about equal.
- It would be nice to have some markup for shell escapes like `< > |` to import or export text.
- Not sure if it worths the effort to support editing of existing files. Ed is the standard text editor.
