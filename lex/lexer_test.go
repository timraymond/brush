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
		{itemText, 0, "This is some plain old *markdown*"},
	}},
	{"other markdown", "This is _other_ markdown", []item{
		{itemText, 0, "This is _other_ markdown"},
	}},
	{"markdown with newlines", "Markdown with \n\n Newlines...", []item{
		{itemText, 0, "Markdown with \n\n Newlines..."},
	}},
	{"markdown with opening braai tag", "The {{", []item{
		{itemText, 0, "The "},
		{itemLeftMeta, 0, "{{"},
		{itemError, 0, "Unexpected end of action"},
	}},
	{"markdown with {{product.name}} braai tag", "The {{product.name}}", []item{
		{itemText, 0, "The "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "product"},
		{itemDot, 0, "."},
		{itemCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
	}},
	{"markdown with multiple {{product.name}} braai tags", "The {{product.name}} is called the {{product.name}}", []item{
		{itemText, 0, "The "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "product"},
		{itemDot, 0, "."},
		{itemCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, " is called the "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "product"},
		{itemDot, 0, "."},
		{itemCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment with arguments", "Here's an awesome attachment {{ attachments(350661) }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemSpace, 0, " "},
		{itemCommand, 0, "attachments"},
		{itemArgumentOpen, 0, "("},
		{itemIdentifier, 0, "350661"},
		{itemArgumentClose, 0, ")"},
		{itemSpace, 0, " "},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment in a callout", "# Editors’ Choice Awards{{callout}}{{ attachments(349807) }}{{/callout}}", []item{
		{itemText, 0, "# Editors’ Choice Awards"},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "callout"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, ""},
		{itemLeftMeta, 0, "{{"},
		{itemSpace, 0, " "},
		{itemCommand, 0, "attachments"},
		{itemArgumentOpen, 0, "("},
		{itemIdentifier, 0, "349807"},
		{itemArgumentClose, 0, ")"},
		{itemSpace, 0, " "},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, ""}, // These are transition points between braai/text
		{itemLeftMeta, 0, "{{"},
		{itemSlash, 0, "/"},
		{itemCommand, 0, "callout"},
		{itemRightMeta, 0, "}}"},
	}},
}

// Lexes the document in the test and returns a slice of items
func collect(t *lexTest) (items []item) {
	lexer := NewLexer(t.input)
	for {
		item := lexer.NextToken()
		items = append(items, item)
		if item.Type == itemEOF || item.Type == itemError {
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
