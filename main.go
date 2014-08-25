package main

import (
	"fmt"
	"github.com/reviewed/brush/lex"
	"io/ioutil"
)

const (
	red         = "\x1b[31m"
	green       = "\x1b[32m"
	brightgreen = "\x1b[1;32m"
	yellow      = "\x1b[33m"
	blue        = "\x1b[34m"
	magenta     = "\x1b[35m"
	reset       = "\x1b[0m"
)

func main() {
	contents, err := ioutil.ReadFile("test.md")
	if err != nil {
		fmt.Println("Unable to read file")
	}
	lexer := lex.NewLexer(string(contents))
	for {
		tok := lexer.NextToken()

		if n := int(tok.Type); n == 11 || n == 13 {
			break
		}
		switch int(tok.Type) {
		case 0:
			fmt.Printf(tok.Value)
		case 1, 2:
			prettyPrint(blue, tok.Value)
		case 3:
			prettyPrint(yellow, tok.Value)
		case 10:
			prettyPrint(magenta, tok.Value)
		case 8, 9:
			prettyPrint(green, tok.Value)
		default:
			fmt.Printf(tok.Value)
		}
	}
}

func prettyPrint(color string, content string) {
	fmt.Printf("%s%s%s", color, content, reset)
}
