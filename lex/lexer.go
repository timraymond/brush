package lex

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	input string    // the string being scanned
	items chan item // holds scanned items
	state stateFn   // current state of the lexer
	start int       // start position in the input of the next token
	pos   int       // current position in the input, will mark the end of the next token
	width int       // width of the last read rune
}

// Represents different types of lexed items.
type itemType int

type stateFn func(*lexer) stateFn

const letters = "abcdefghijklmnopqrstuvwxyz"
const alphaNum = "0123456789abcdefghijklmnopqrstuvwxyz"

const eof = -1

const (
	itemText itemType = iota
	itemLeftMeta
	itemRightMeta
	itemCommand
	itemDot
	itemSlash
	itemArgumentOpen
	itemArgumentClose
	itemParenthesizedArgument
	itemQuotedArgument
	itemDotCommand
	itemIdentifier
	itemEOF
	itemSpace
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
func NewLexer(document string) *lexer {
	return &lexer{input: document, state: lexText, items: make(chan item, 2)}
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
	l.emit(itemLeftMeta)
	return lexInsideAction
}

func lexRightMeta(l *lexer) stateFn {
	l.pos += len("}")
	l.emit(itemRightMeta)
	return lexText
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
		return lexCommand
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
	case r == '/':
		return lexCloser
	default:
		return l.errorf("Unexpected character %#U", r)
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
	l.acceptRun(alphaNum)
	if l.peek() == opener {
		l.emit(itemQuotedArgument)
		l.next()
		l.ignore()
	} else {
		return l.errorf("Unbalanced quoting in argument")
	}
	return lexInsideAction
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

func lexCloser(l *lexer) stateFn {
	l.emit(itemSlash)
	return lexCommand
}

func lexCommand(l *lexer) stateFn {
	l.acceptRun(letters)
	l.emit(itemCommand)
	return lexInsideAction
}

func lexArgumentOpen(l *lexer) stateFn {
	l.emit(itemArgumentOpen)
	return lexArgument
}

func lexArgumentClose(l *lexer) stateFn {
	r := l.next()
	if r == ')' || r == '\'' {
		l.emit(itemArgumentClose)
	} else {
		return l.errorf("Unexpected character in argument: %#U", r)
	}
	return lexInsideAction
}

func lexArgument(l *lexer) stateFn {
	l.acceptRun(alphaNum)
	l.emit(itemIdentifier)
	return lexArgumentClose
}

func isSpace(input rune) bool {
	if input == ' ' || input == '\t' {
		return true
	} else {
		return false
	}
}
