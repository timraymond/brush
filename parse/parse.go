package parse

import "fmt"
import "strconv"

// A Tree holds all of the parsing state necessary to transform a document into
// an AST
type Tree struct {
	lexer      *lexer // the lexer which is the source of tokens
	Error      error  // the last returned error
	ParseName  string // the name of the document being parsed
	token      item   // maintains one token lookahead
	peekCount  int    // count of how many tokens of lookahead we have
	lastCol    int    // column of the last item in the lookahead buffer
	blockLevel int    // nesting level of block tags
}

func (t *Tree) formatPos() string {
	if t.peekCount == 0 {
		return t.ParseName + ":" + strconv.Itoa(t.lexer.lineNumber()) + ":" + strconv.Itoa(t.lexer.column()) + ": "
	} else {
		return t.ParseName + ":" + strconv.Itoa(t.lexer.lineNumber()) + ":" + strconv.Itoa(t.lastCol) + ": "
	}
}

// Parse creates an AST from the document that the parser was initialized with.
func (t *Tree) Parse() (root Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	root = t.document()
	if t.Error != nil {
		return nil, t.Error
	} else {
		return root, nil
	}
}

// DOCUMENT -> itemText DOCUMENT
//           | BRAAI DOCUMENT
//           | ε
func (t *Tree) document() Node {
	root := &DocumentNode{}
	root.NodeList = make([]Node, 0)
	for {
		tok := t.lexer.NextToken()
		switch tok.Type {
		case itemText:
			root.NodeList = append(root.NodeList, &TextNode{[]byte(tok.Value)})
		case itemEOF:
			return root
		case itemError:
			t.Error = fmt.Errorf(tok.Value)
			return root
		case itemCloser:
			if t.blockLevel > 0 {
				return root
			} else {
				t.Error = fmt.Errorf("Unexpected closing tag at at %d", tok.Pos)
			}
		case itemLeftMeta:
			root.NodeList = append(root.NodeList, t.blockOrRegular())
			if t.Error != nil {
				return root
			}
		default:
			t.Error = fmt.Errorf("Unexpected token %s at %d", tok.Type, tok.Pos)
			return root
		}
	}
	return root
}

// BLOCK_OR_REGULAR -> itemBlock itemRightMeta DOCUMENT itemCloser itemBlock itemRightMeta
//                    | REGULAR
func (t *Tree) blockOrRegular() Node {
	tok := t.expectOneOf(itemIdentifier, itemBlock)
	if tok.Type == itemIdentifier {
		t.backup()
		return t.braaiTag()
	} else if tok.Type == itemBlock {
		t.backup()
		return t.blockTag()
	} else if tok.Type == itemError {
		t.Error = fmt.Errorf("Lexical Error while parsing BLOCK_OR_REGULAR: %s", tok.Value)
		return &BraaiTagNode{}
	} else {
		t.Error = fmt.Errorf("Unexpected token %s at %d", tok.Type, tok.Pos)
		return &BraaiTagNode{}
	}
}

// REGULAR -> itemIdent DOTCOMMANDS ARG_LIST MODIFIERS itemRightMeta
func (t *Tree) braaiTag() Node {
	posFormat := t.formatPos()
	ident := t.expect(itemIdentifier, "braai tag")
	tok := t.next()
	var arguments []string = make([]string, 0)
	switch tok.Type {
	case itemParenthesizedArgument, itemBracketedArgument:
		arguments = append(arguments, tok.Value)
	default:
		t.backup()
	}
	dotCommands := t.dotCommands()
	for _, arg := range t.argumentList() {
		arguments = append(arguments, arg)
	}
	attrs := t.attributes()
	t.expect(itemRightMeta, "braai tag")
	return &BraaiTagNode{ident.Value, dotCommands, arguments, attrs, posFormat}
}

func (t *Tree) argumentList() (arguments []string) {
	for {
		tok := t.next()
		if tok.Type != itemQuotedArgument {
			t.backup()
			break
		}
		arguments = append(arguments, tok.Value)
	}
	return
}

func (t *Tree) attributes() map[string]string {
	const context string = "attribute list"
	attrs := make(map[string]string)
	for {
		if t.next().Type == itemRightMeta {
			t.backup()
			return attrs
		}
		t.backup()
		key := t.expect(itemIdentifier, context)
		t.expect(itemAssign, context)
		value := t.expect(itemQuotedArgument, context)
		if t.Error != nil {
			return attrs
		}
		attrs[key.Value] = value.Value
	}
	return attrs
}

func (t *Tree) blockTag() Node {
	const context string = "block tag"
	tok := t.expect(itemBlock, context)
	t.expect(itemRightMeta, context)
	t.blockLevel++
	body := t.document()
	end_tok := t.expect(itemBlock, context)
	t.expect(itemRightMeta, context)
	if tok.Value != end_tok.Value {
		t.Error = fmt.Errorf("Mismatched block tag, opener: %s, closer: %s", tok.Value, end_tok.Value)
	}
	t.blockLevel--
	return &BlockTagNode{tok.Value, body}
}

// DOTCOMMANDS -> itemDotCommand SINGLE_ARGS DOTCOMMANDS | ε
func (t *Tree) dotCommands() (dotCommands []DotCommandNode) {
	for {
		tok := t.next()
		if tok.Type != itemDotCommand {
			t.backup()
			break
		} else {
			argument := t.singleArgument()
			dotCommands = append(dotCommands, DotCommandNode{tok.Value, argument})
		}
	}
	return dotCommands
}

func (t *Tree) singleArgument() Node {
	tok := t.next()
	switch tok.Type {
	case itemParenthesizedArgument, itemBracketedArgument:
		return &SingleArgumentNode{tok.Value}
	case itemDotCommand, itemRightMeta:
		t.backup()
		return nil
	default:
		t.Error = fmt.Errorf("Unexpected token %s in single argument at %d", tok.Value, tok.Pos)
		return nil
	}
}

func (t *Tree) next() item {
	if t.peekCount > 1 {
		panic("Parser lookahead overflow")
	} else if t.peekCount == 0 {
		t.lastCol = t.lexer.column()
		t.token = t.lexer.NextToken()
	} else {
		t.peekCount--
	}
	return t.token
}

func (t *Tree) backup() {
	t.peekCount++
}

func (t *Tree) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s:%d:%d: %s", t.ParseName, t.lexer.lineNumber(), t.lexer.column(), format)
	panic(fmt.Errorf(format, args...))
}

func (t *Tree) expectOneOf(expected ...itemType) item {
	tok := t.next()
	for _, expectedTok := range expected {
		if tok.Type == expectedTok {
			return tok
		}
	}
	t.backup()
	t.Error = fmt.Errorf("Unexpected %s at position %d", tok.Value, tok.Pos)
	return tok
}

func (t *Tree) expect(expected itemType, context string) item {
	tok := t.next()
	if tok.Type != expected {
		t.backup()
		if tok.Type == itemError {
			t.errorf("Lexical Error - %s", tok.Value)
		} else {
			t.errorf("Unexpected %d, expected to see %d", tok.Value, expected)
		}
		//t.Error = fmt.Errorf("Unexpected %s in %s at position %d. Expected %d, saw %d", tok.Value, context, tok.Pos, expected, tok.Type)
	}
	return tok
}

// New returns a *Tree which is initialized with a lexer so that parsing can
// proceed immediately
func New(name string, input string, blockTags []string) *Tree {
	return &Tree{ParseName: name, lexer: NewLexer(input, blockTags), Error: nil}
}
