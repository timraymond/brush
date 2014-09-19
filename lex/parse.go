package lex

import "fmt"

type Tree struct {
  lexer *lexer
  Error error
  token item
  peekCount int
  blockLevel int
}

func (t *Tree) Parse() (Node, error) {
  root := t.document()
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
      t.Error = fmt.Errorf("Unexpected token %d at %d", tok.Type, tok.Pos)
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
  } else {
    t.Error = fmt.Errorf("Unexpected token %d at %d", tok.Type, tok.Pos)
    return &BraaiTagNode{}
  }
}

// REGULAR -> itemIdent DOTCOMMANDS ARG_LIST MODIFIERS itemRightMeta
func (t *Tree) braaiTag() Node {
  tok := t.expect(itemIdentifier, "braai tag")
  dotCommands := t.dotCommands()
  attrs := t.attributes()
  t.expect(itemRightMeta, "braai tag")
  return &BraaiTagNode{tok.Value, dotCommands, attrs}
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
func (t *Tree) dotCommands() (dotCommands []Node) {
  for {
    tok := t.next()
    if tok.Type != itemDotCommand {
      t.backup()
      break
    } else {
      argument := t.singleArgument()
      dotCommands = append(dotCommands, &DotCommandNode{tok.Value, argument})
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
  if t.peekCount == 0 {
    t.token = t.lexer.NextToken()
  } else {
    t.peekCount--
  }
  return t.token
}

func (t *Tree) backup() {
  t.peekCount++
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
    t.Error = fmt.Errorf("Unexpected %s in %s at position %d. Expected %d, saw %d", tok.Value, context, tok.Pos, expected, tok.Type)
  }
  return tok
}

func New(input string) *Tree {
  return &Tree{lexer: NewLexer(input, []string{"callout"}), Error: nil}
}
