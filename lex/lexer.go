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

func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
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
		l.emit(itemDot)
		return lexInsideAction
	case r == '(' || r == '\'':
		return lexArgumentOpen
	case r == '/':
		return lexCloser
	default:
		return l.errorf("Unexpected character %#U", r)
	}
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
