package lex

import (
	"testing"
)

type lexTest struct {
	name  string
	input string
	items []item
}

var lexTests = []lexTest{
	{"plain", "This is some plain old *markdown*", []item{
		{itemText, "This is some plain old *markdown*"},
	}},
	{"other markdown", "This is _other_ markdown", []item{
		{itemText, "This is _other_ markdown"},
	}},
	{"markdown with newlines", "Markdown with \n\n Newlines...", []item{
		{itemText, "Markdown with \n\n Newlines..."},
	}},
}

func TestLexing(t *testing.T) {
	for _, test := range lexTests {
		lexer := NewLexer(test.input)
		for _, expectedItem := range test.items {
			item, err := lexer.NextToken()
			if err != nil {
				t.Error(err)
			}
			if item.Value != expectedItem.Value {
				t.Errorf("LexTest: %s, Unexpected Value. Expected \"%s\" to eqaul \"%s\"", test.name, item.Value, expectedItem.Value)
			}
			if item.Type != expectedItem.Type {
				t.Errorf("LexTest: %s, Unexpected Item. Expected \"%s\" to eqaul \"%s\"", test.name, item.Type, expectedItem.Type)
			}
		}
	}
}
