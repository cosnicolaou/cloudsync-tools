package main

import (
	"fmt"
	"strings"

	"github.com/buger/goterm"
)

type formatter struct {
	width   int
	output  *strings.Builder
	lineLen int
}

func newFormatter() *formatter {
	return &formatter{width: goterm.Width(), output: &strings.Builder{}}
}

func (f *formatter) append(s string) {
	if f.lineLen+len(s)+2 > f.width {
		f.output.WriteRune('\n')
		f.lineLen = 0
	}
	f.output.WriteString(s)
	f.output.WriteString("  ")
	f.lineLen += len(s) + 2
}

func (f *formatter) flush() {
	fmt.Println(f.output.String())
	f.output.Reset()
}
