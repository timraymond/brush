package lex

import "strings"

const (
	red         = "\x1b[31m"
	green       = "\x1b[32m"
	brightgreen = "\x1b[1;32m"
	yellow      = "\x1b[33m"
	blue        = "\x1b[34m"
	magenta     = "\x1b[35m"
	reset       = "\x1b[0m"
)

type Node interface{
  String() string
}

type DocumentNode struct {
  NodeList []Node
}

func (d *DocumentNode) String() string {
  var parts = make([]string, 0)
  for _, node := range d.NodeList {
    parts = append(parts, node.String())
  }
  return strings.Join(parts, "")
}

type BlockTagNode struct {
  Name string
  Subtree Node
}

func (b *BlockTagNode) String() string {
  return magenta + "{{ " + b.Name + " }} " + reset + b.Subtree.String() + magenta + "{{/ " + b.Name + " }}" + reset
}

type TextNode struct {
  Text []byte
}

func (t *TextNode) String() string {
  return string(t.Text)
}

type BraaiTagNode struct {
  Text string
  DotCommands []Node
  Attributes map[string]string
}

func (b *BraaiTagNode) String() string {
  var parts = make([]string, 2)
  parts[0] = red + "{{ "
  parts[1] = b.Text
  for _, cmd := range b.DotCommands {
    parts = append(parts, cmd.String())
  }
  for k, v := range b.Attributes {
    parts = append(parts, yellow + " " + k + "=" + "\"" + v + "\" " + reset)
  }
  parts = append(parts, red + " }}" + reset)
  return strings.Join(parts, "")
}

type SingleArgumentNode struct {
  Text string
}

func (s *SingleArgumentNode) String() string {
  return blue + "( " + s.Text + " )" + reset
}

type DotCommandNode struct {
  Text string
  Argument Node
}

func (d *DotCommandNode) String() string {
  if d.Argument != nil {
    return green + "." + d.Text + d.Argument.String() + reset
  } else {
    return green + "." + d.Text + reset
  }
}
