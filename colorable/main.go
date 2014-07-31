package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-colorable"
)

func main() {
	w := colorable.NewColorableStdout()
	msg := "Rainbow"

	if len(os.Args) > 1 {
		msg = os.Args[1]
	}

	rainbowize(msg, w)
}

func rainbowize(s string, w io.Writer) {
	cc := 8 //colors
	rc := utf8.RuneCountInString(s)
	offset := 0
	for {
		//clear
		fmt.Fprintf(w, strings.Repeat("\b", rc))

		for i, r := range s {
			fmt.Fprintf(w, "\x1b[3%d;1m%s", (i+offset)%cc, string(r))
		}
		offset++

		//wait
		time.Sleep(250 * time.Millisecond)
	}
}
