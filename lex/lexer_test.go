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
	{"markdown with opening braai tag", "The {{", []item{
		{itemText, "The "},
		{itemLeftMeta, "{{"},
	}},
	{"markdown with {{product.name}} braai tag", "The {{product.name}}", []item{
		{itemText, "The "},
		{itemLeftMeta, "{{"},
		{itemProductName, "product.name"},
		{itemRightMeta, "}}"},
	}},
	{"markdown with multiple {{product.name}} braai tags", "The {{product.name}} is called the {{product.name}}", []item{
		{itemText, "The "},
		{itemLeftMeta, "{{"},
		{itemProductName, "product.name"},
		{itemRightMeta, "}}"},
		{itemText, " is called the "},
		{itemLeftMeta, "{{"},
		{itemProductName, "product.name"},
		{itemRightMeta, "}}"},
	}},
	{"attachment with arguments", "Here's an awesome attachment {{attachments(350661)}}", []item{
		{itemText, "Here's an awesome attachment "},
		{itemLeftMeta, "{{"},
		{itemAttachment, "attachments"},
		{itemArgumentOpen, "("},
		{itemIdentifier, "350661"},
		{itemArgumentClose, ")"},
		{itemRightMeta, "}}"},
	}},
}

// Lexes the document in the test and returns a slice of items
func collect(t *lexTest) (items []item) {
	lexer := NewLexer(t.input)
	for {
		item := lexer.NextToken()
		items = append(items, item)
		if item.Type == itemEOF {
			break
		}
	}
	return
}

func TestLexing(t *testing.T) {
	for _, test := range lexTests {
		actualItems := collect(&test)

		for idx, expected := range test.items {
			actual := actualItems[idx]
			if expected.Value != actual.Value {
				t.Errorf("%s\n\tExpected \"%s\" to equal \"%s\"", test.name, actual.Value, expected.Value)
			}
			if expected.Type != actual.Type {
				t.Errorf("%s\n\tExpected \"%s\" to equal \"%s\"", test.name, actual.Type, expected.Type)
			}
		}
	}
}
