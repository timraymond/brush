package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	input    string    // the string being scanned
	items    chan item // holds scanned items
	blockIds []string  // identifiers which should be treated as block elments
	state    stateFn   // current state of the lexer
	start    int       // start position in the input of the next token
	pos      int       // current position in the input, will mark the end of the next token
	width    int       // width of the last read rune
}

// Represents different types of lexed items.
type itemType int

type stateFn func(*lexer) stateFn

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_"
const alphaNum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ "
const filename = "-._()/,&é0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ "

const eof = -1

const (
	itemText      itemType = iota // an unprocessed block of opaque text
	itemLeftMeta                  // the beginning of a braai tag
	itemRightMeta                 // the end of a braai tag
	itemBlock                     // an identifier specified as a block tag
	itemCloser                    // a block identifier prefixed by a slash
	itemParenthesizedArgument
	itemQuotedArgument
	itemBracketedArgument
	itemDotCommand
	itemAssign
	itemIdentifier
	itemEOF
	itemError
)

// Represents a scanned item. Includes the string value encompassing the item.
// Also known as a Lexeme or Token in other lexer literature
type item struct {
	Type  itemType
	Pos   int
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

// Provides a new lexer for the given document
func NewLexer(document string, blockIds []string) *lexer {
	return &lexer{input: document, state: lexText, items: make(chan item, 2), blockIds: blockIds}
}

// emits a new item into the lexer's items channel
func (self *lexer) emit(t itemType) {
	self.items <- item{t, self.start, self.input[self.start:self.pos]}
	self.start = self.pos
}

func (self *lexer) backup() {
	self.pos -= self.width
}

func (self *lexer) next() rune {
	if int(self.pos) >= len(self.input) {
		return eof
	}
	r, w := utf8.DecodeRuneInString(self.input[self.pos:])
	self.width = w
	self.pos += w
	return r
}

func (self *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, self.next()) >= 0 {
		return true
	}
	self.backup()
	return false
}

func (self *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, self.next()) >= 0 {
	}
	self.backup()
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

func (l *lexer) column() int {
	lastNewline := strings.LastIndex(l.input[:l.pos], "\n")
	if lastNewline == -1 {
		lastNewline = 0
	}
	return utf8.RuneCountInString(l.input[lastNewline:l.start])
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "{{") {
			l.emit(itemText)
			return lexLeftMeta
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func lexLeftMeta(l *lexer) stateFn {
	l.pos += len("{{")
	if tok := l.next(); tok == '/' {
		l.emit(itemCloser)
	} else if tok == eof {
		l.emit(itemLeftMeta)
	} else {
		l.backup()
		l.emit(itemLeftMeta)
	}
	return lexInsideAction
}

func lexRightMeta(l *lexer) stateFn {
	if l.accept("}}") {
		l.emit(itemRightMeta)
		return lexText
	} else {
		return l.errorf("Malformed end of Braai tag, should be }}")
	}
}

// Braai ignores all whitespace within braai tags
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexInsideAction
}

func lexInsideAction(l *lexer) stateFn {
	switch r := l.next(); {
	case unicode.IsLetter(r):
		return lexIdentifier
	case isSpace(r):
		return lexSpace
	case r == eof:
		return l.errorf("Unexpected end of action")
	case r == '}':
		return lexRightMeta
	case r == '.':
		return lexDotCommand
	case r == '(':
		return lexParenthesizedArgument
	case r == '\'' || r == '"':
		return lexQuotedArgument
	case r == '=':
		l.emit(itemAssign)
		if r = l.next(); r == '\'' || r == '"' {
			return lexQuotedArgument
		} else if r == 't' || r == 'f' {
			return lexBoolean
		} else {
			return l.errorf("Malformed modifier")
		}
	case r == '[':
		return lexBracketedArgument
	case r == ',' || r == '\n':
		l.ignore()
		return lexInsideAction
	default:
		return l.errorf("Unexpected character %#U", r)
	}
}

func lexBoolean(l *lexer) stateFn {
	l.backup()
	if l.accept("t") {
		for i := 0; i < 3; i++ {
			l.next()
		}
		if boolean := l.input[l.start:l.pos]; boolean == "true" {
			l.emit(itemQuotedArgument)
		} else {
			l.errorf("Expected boolean true, saw %s", boolean)
		}
	} else if l.accept("f") {
		for i := 0; i < 4; i++ {
			l.next()
		}
		if boolean := l.input[l.start:l.pos]; boolean == "false" {
			l.emit(itemQuotedArgument)
		} else {
			l.errorf("Expected boolean false, saw %s", boolean)
		}
	}

	return lexInsideAction
}

// Lexes arguments of the form ['A Tag'] or ["Another Tag"]. Enforces correct
// balancing of quotation marks
func lexBracketedArgument(l *lexer) stateFn {
	if r := l.next(); r == '\'' || r == '"' {
		l.ignore() // the parser is uninterested in quotations
		for {
			switch l.next() {
			case r:
				if l.peek() == ']' {
					l.backup()
					l.emit(itemBracketedArgument)
					l.next()   // grab the closing quote
					l.next()   // ... and the closing bracket
					l.ignore() // ... and throw them away
					return lexInsideAction
				} else {
					return l.errorf("Malformed bracketed argument, expected quote")
				}
			case eof:
				return l.errorf("Unterminated bracketed argument")
			}
		}
	} else {
		return l.errorf("Malformed bracketed argument, expected quote")
	}
}

func lexDotCommand(l *lexer) stateFn {
	l.ignore() // ignore the dot, we know it's there
	l.acceptRun(letters)
	l.emit(itemDotCommand)
	return lexInsideAction
}

func lexQuotedArgument(l *lexer) stateFn {
	l.backup()
	opener := l.next()
	l.ignore()
	for {
		switch l.next() {
		case opener:
			l.backup()
			l.emit(itemQuotedArgument)
			l.next()
			l.ignore()
			return lexInsideAction
		case eof:
			return l.errorf("Unterminated quoted argument")
		}
	}
}

// Returns a itemParenthesizedArgument token sans parentheses
func lexParenthesizedArgument(l *lexer) stateFn {
	l.ignore()
	l.acceptRun(alphaNum)
	if l.peek() == ')' {
		l.emit(itemParenthesizedArgument)
		l.next()
		l.ignore()
	} else {
		return l.errorf("Missing closing parenthesis on argument")
	}
	return lexInsideAction
}

func lexIdentifier(l *lexer) stateFn {
	l.acceptRun(letters)
	id := string([]byte(l.input)[l.start:l.pos])
	for _, blockId := range l.blockIds {
		if blockId == id {
			l.emit(itemBlock)
			return lexInsideAction
		}
	}
	l.emit(itemIdentifier)
	return lexInsideAction
}

func isSpace(input rune) bool {
	if input == ' ' || input == '\t' {
		return true
	} else {
		return false
	}
}
