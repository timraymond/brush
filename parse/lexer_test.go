package parse

import "testing"

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
		{itemIdentifier, 0, "product"},
		{itemDotCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
	}},
	{"markdown with multiple {{product.name}} braai tags", "The {{product.name}} is called the {{product.name}}", []item{
		{itemText, 0, "The "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "product"},
		{itemDotCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, " is called the "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "product"},
		{itemDotCommand, 0, "name"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment with arguments", "Here's an awesome attachment {{ attachments(350661) }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemRightMeta, 0, "}}"},
	}},
	{"attachment in a callout", "# Editors’ Choice Awards{{callout}}{{ attachments(349807) }}{{/callout}}", []item{
		{itemText, 0, "# Editors’ Choice Awards"},
		{itemLeftMeta, 0, "{{"},
		{itemBlock, 0, "callout"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, ""},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemParenthesizedArgument, 0, "349807"},
		{itemRightMeta, 0, "}}"},
		{itemText, 0, ""}, // These are transition points between braai/text
		{itemCloser, 0, "{{/"},
		{itemBlock, 0, "callout"},
		{itemRightMeta, 0, "}}"},
	}},
	{"brightcove", "Behold, a video: {{ brightcove '1234' }}", []item{
		{itemText, 0, "Behold, a video: "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "brightcove"},
		{itemQuotedArgument, 0, "1234"},
		{itemRightMeta, 0, "}}"},
	}},
	{"brightcove with double quotes", `Behold, a video: {{ brightcove "1234" }}`, []item{
		{itemText, 0, "Behold, a video: "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "brightcove"},
		{itemQuotedArgument, 0, "1234"},
		{itemRightMeta, 0, "}}"},
	}},
	{"quoted argument with unbalanced quotes", `Behold, a video: {{ brightcove "1234' }}`, []item{
		{itemText, 0, "Behold, a video: "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "brightcove"},
		{itemError, 0, "Unterminated quoted argument"},
	}},
	{"attachment with popup", "Here's an awesome attachment {{ attachments(350661).popup }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemRightMeta, 0, "}}"},
	}},
	{"unbalanced parentheses", "Here's an awesome attachment {{ attachments(350661 }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemError, 0, "Missing closing parenthesis on argument"},
	}},
	{"attachment with modifiers", "Here's an awesome attachment {{ attachments(350661).popup big='true' }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemIdentifier, 0, "big"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "true"},
		{itemRightMeta, 0, "}}"},
	}},
	{"reports malformed modifiers properly", "Here's an awesome attachment {{ attachments(350661).popup big= }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemIdentifier, 0, "big"},
		{itemAssign, 0, "="},
		{itemError, 0, "Malformed modifier"},
	}},
	{"reports unexpected characters properly", "Here's an awesome attachment {{ ? }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemError, 0, "Unexpected character U+003F '?'"},
	}},
	{"multiple comma-separated modifiers", "Here's an awesome attachment {{ attachments(350661).popup big='true', bacon=\"yes\" }}", []item{
		{itemText, 0, "Here's an awesome attachment "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "attachments"},
		{itemParenthesizedArgument, 0, "350661"},
		{itemDotCommand, 0, "popup"},
		{itemIdentifier, 0, "big"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "true"},
		{itemIdentifier, 0, "bacon"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "yes"},
		{itemRightMeta, 0, "}}"},
	}},
	{"ignore commas inside action", "Some plain text {{ photo_gallery 'Foo', \"Bar\"}}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "photo_gallery"},
		{itemQuotedArgument, 0, "Foo"},
		{itemQuotedArgument, 0, "Bar"},
		{itemRightMeta, 0, "}}"},
	}},
	{"unbalanced opening and closing meta", "Some plain text {{ photo_gallery 'Foo', \"Bar\"}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "photo_gallery"},
		{itemQuotedArgument, 0, "Foo"},
		{itemQuotedArgument, 0, "Bar"},
		{itemError, 0, "Malformed end of Braai tag, should be }}"},
	}},
	{"ignore newlines in Braai tags", "Some plain text {{ photo_gallery 'Foo', \n\"Bar\"}}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "photo_gallery"},
		{itemQuotedArgument, 0, "Foo"},
		{itemQuotedArgument, 0, "Bar"},
		{itemRightMeta, 0, "}}"},
	}},
	{"bracketed arguments", "Some plain text {{ article.attachments[\"The Thing\"] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemBracketedArgument, 0, "The Thing"},
		{itemRightMeta, 0, "}}"},
	}},
	{"bracketed arguments with a real file name", "Some plain text {{ article.attachments['Electrolux-EIDW5905JS-FrontClosed2.jpg'] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemBracketedArgument, 0, "Electrolux-EIDW5905JS-FrontClosed2.jpg"},
		{itemRightMeta, 0, "}}"},
	}},
	{"quoted arguments with special characters", "Some plain text {{ youtube 'fg_12345 / (turtles & unicorns in a café, of course)' }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "youtube"},
		{itemQuotedArgument, 0, "fg_12345 / (turtles & unicorns in a café, of course)"},
		{itemRightMeta, 0, "}}"},
	}},
	{"bracketed arguments with special characters", "Some plain text {{ article.attachments['café-photos-&-first-floor-/-second-floor-(ね)).jpg']}}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemBracketedArgument, 0, "café-photos-&-first-floor-/-second-floor-(ね)).jpg"},
		{itemRightMeta, 0, "}}"},
	}},
	{"Unbalanced quoting in bracketed arguments", "Some plain text {{ article.attachments[\"The Thing'] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemError, 0, "Unterminated bracketed argument"},
	}},
	{"Unexpected termination of bracketed argument", "Some plain text {{ article.attachments['The Thing's things'] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemError, 0, "Malformed bracketed argument, expected quote"},
	}},
	{"Malformed bracketed argument", "Some plain text {{ article.attachments[] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemError, 0, "Malformed bracketed argument, expected quote"},
	}},
	{"ignore arbitrary number of spaces", "Some plain text {{         article.attachments[\"The Thing\"] }}", []item{
		{itemText, 0, "Some plain text "},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "article"},
		{itemDotCommand, 0, "attachments"},
		{itemBracketedArgument, 0, "The Thing"},
		{itemRightMeta, 0, "}}"},
	}},
	{"handle capital letters in modifier name", "{{comparison_bars title=\"Video: Low Light Noise Score Comparison\", attribute=\"Video Low Light Noise Score\", comps=\"video\", xLabel=\"Video Low Light Noise Score\"}}", []item{
		{itemText, 0, ""},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "comparison_bars"},
		{itemIdentifier, 0, "title"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "Video: Low Light Noise Score Comparison"},
		{itemIdentifier, 0, "attribute"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "Video Low Light Noise Score"},
		{itemIdentifier, 0, "comps"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "video"},
		{itemIdentifier, 0, "xLabel"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "Video Low Light Noise Score"},
		{itemRightMeta, 0, "}}"},
	}},
	{"handle boolean modifiers", "{{product.attachments['FI Handling Photo 2'] include_caption=true}}", []item{
		{itemText, 0, ""},
		{itemLeftMeta, 0, "{{"},
		{itemIdentifier, 0, "product"},
		{itemDotCommand, 0, "attachments"},
		{itemBracketedArgument, 0, "FI Handling Photo 2"},
		{itemIdentifier, 0, "include_caption"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "true"},
		{itemRightMeta, 0, "}}"},
	}},
  {"handle colons and dots in modifier names", "{{ product_shelf max:msrp=\"800\" category__slug=\"foo\" }}", []item{
    {itemText, 0, ""},
    {itemLeftMeta, 0, "{{"},
    {itemIdentifier, 0, "product_shelf"},
    {itemIdentifier, 0, "max:msrp"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "800"},
    {itemIdentifier, 0, "category__slug"},
		{itemAssign, 0, "="},
		{itemQuotedArgument, 0, "foo"},
		{itemRightMeta, 0, "}}"},
  }},
}

// Lexes the document in the test and returns a slice of items
func collect(t *lexTest) (items []item) {
	blockElements := []string{"callout"}
	lexer := NewLexer(t.input, blockElements)
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
