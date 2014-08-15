package lex

import "strings"

type lexer struct {
	input string    // the string being scanned
	items chan item // holds scanned items
	state stateFn   // current state of the lexer
	start int       // start position in the input
	pos   int       // current position in the input
}

// Represents different types of lexed items.
type itemType int

type stateFn func(*lexer) stateFn

const (
	itemText itemType = iota
	itemLeftMeta
	itemRightMeta
	itemProductName
	itemEOF
)

var keywords = map[string]itemType{
	"product.name": itemProductName,
}

// Represents a scanned item. Includes the string value encompassing the item.
// Also known as a Lexeme or Token in other lexer literature
type item struct {
	Type  itemType
	Value string
}

func (self *lexer) NextToken() item {
	for {
		select {
		case item := <-self.items:
			return item
		default:
			self.state = self.state(self)
		}
	}
}

// emits a new item into the lexer's items channel
func (self *lexer) emit(t itemType) {
	self.items <- item{t, self.input[self.start:self.pos]}
	self.start = self.pos
}

// Provides a new lexer for the given document
func NewLexer(document string) *lexer {
	return &lexer{input: document, state: lexText, items: make(chan item, 2)}
}

func lexText(l *lexer) stateFn {
	for {
		if l.pos == len(l.input) {
			l.emit(itemText)
			return lexEOF
		}
		if strings.HasPrefix(l.input[l.pos:], "{{") {
			l.emit(itemText)
			return lexLeftMeta
		} else {
			l.pos += 1
		}
	}
	l.emit(itemText)
	return lexEOF
}

func lexLeftMeta(l *lexer) stateFn {
	l.pos += 2
	l.emit(itemLeftMeta)
	return lexInsideAction
}

func lexRightMeta(l *lexer) stateFn {
	l.pos += 2
	l.emit(itemRightMeta)
	return lexText
}

func lexInsideAction(l *lexer) stateFn {
	for {
		if l.pos == len(l.input) {
			l.emit(itemText)
			return lexEOF
		}
		if strings.HasPrefix(l.input[l.pos:], "}}") {
			l.emit(keywords[l.input[l.start:l.pos]])
			return lexRightMeta
		} else {
			l.pos += 1
		}
	}
}

func lexEOF(l *lexer) stateFn {
	l.emit(itemEOF)
	return nil
}
