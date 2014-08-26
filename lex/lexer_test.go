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
		{itemDotCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
	}},
	{"markdown with multiple {{product.name}} braai tags", "The {{product.name}} is called the {{product.name}}", []item{
		{itemText, 0, "The "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "product"},
		{itemDotCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, " is called the "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "product"},
		{itemDotCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment with arguments", "Here's an awesome attachment {{ attachments(350661) }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment in a callout", "# Editors’ Choice Awards{{callout}}{{ attachments(349807) }}{{/callout}}", []item{
		{itemText, 0, "# Editors’ Choice Awards"},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "callout"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, ""},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "attachments"},
		{itemParenthesizedArgument, 0, "349807"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, ""}, // These are transition points between braai/text
		{itemLeftMeta, 0, "{{"},
		{itemSlash, 0, "/"},
		{itemCommand, 0, "callout"},
		{itemRightMeta, 0, "}}"},
	}},
	{"brightcove", "Behold, a video: {{ brightcove '1234' }}", []item{
		{itemText, 0, "Behold, a video: "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "brightcove"},
		{itemQuotedArgument, 0, "1234"},
		{itemRightMeta, 0, "}}"},
	}},
	{"brightcove with double quotes", `Behold, a video: {{ brightcove "1234" }}`, []item{
		{itemText, 0, "Behold, a video: "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "brightcove"},
		{itemQuotedArgument, 0, "1234"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment with popup", "Here's an awesome attachment {{ attachments(350661).popup }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment with modifiers", "Here's an awesome attachment {{ attachments(350661).popup big='true' }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemCommand, 0, "big"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "true"},
		{itemRightMeta, 0, "}}"},
	}},
	{"multiple comma-separated modifiers", "Here's an awesome attachment {{ attachments(350661).popup big='true', bacon=\"yes\" }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemCommand, 0, "big"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "true"},
		{itemCommand, 0, "bacon"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "yes"},
		{itemRightMeta, 0, "}}"},
	}},
	{"ignore commas inside action", "Some plain text {{ photo_gallery 'Foo', \"Bar\"}}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "photo_gallery"},
		{itemQuotedArgument, 0, "Foo"},
		{itemQuotedArgument, 0, "Bar"},
		{itemRightMeta, 0, "}}"},
	}},
	{"ignore newlines in Braai tags", "Some plain text {{ photo_gallery 'Foo', \n\"Bar\"}}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "photo_gallery"},
		{itemQuotedArgument, 0, "Foo"},
		{itemQuotedArgument, 0, "Bar"},
		{itemRightMeta, 0, "}}"},
	}},
	{"bracketed arguments", "Some plain text {{ article.attachments[\"The Thing\"] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemCommand, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemBracketedArgument, 0, "The Thing"},
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
			// This will catch panics and print something useful in the
			// test output instead
			if actual.Type == itemError && expected.Type != itemError {
				t.Errorf("%s:\n\tLexical Error: %s. Location: %d", test.name, actual.Value, actual.Pos)
				break
			}
			if expected.Value != actual.Value {
				t.Errorf("%s:\n\tExpected \"%s\" to equal \"%s\"", test.name, actual.Value, expected.Value)
			}
			if expected.Type != actual.Type {
				t.Errorf("%s:\n\tExpected \"%s\" to equal \"%s\"", test.name, actual.Type, expected.Type)
			}
		}
	}
}
