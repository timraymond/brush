package lex

type lexer struct {
	input string
}

type itemType int

const (
	itemText itemType = iota
  itemLeftMeta
  itemRightMeta
  itemProductName
)

type item struct {
	Type  itemType
	Value string
}

func (self *lexer) NextToken() (item, error) {
	return item{itemText, self.input}, nil
}

func NewLexer(document string) *lexer {
	return &lexer{input: document}
}
