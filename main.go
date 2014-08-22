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

		if n := int(tok.Type); n == 9 || n == 11 {
			break
		}
		switch int(tok.Type) {
		case 0:
			fmt.Printf(tok.Value)
		case 1, 2:
			fmt.Printf("%s%s%s", blue, tok.Value, reset)
		case 3:
			fmt.Printf("%s%s%s", yellow, tok.Value, reset)
		case 8:
			fmt.Printf("%s%s%s", magenta, tok.Value, reset)
    case 6, 7:
			fmt.Printf("%s%s%s", green, tok.Value, reset)
		default:
			fmt.Printf(tok.Value)
		}
	}
}
