package parse

import (
  "fmt"
  "strings"

  "github.com/russross/blackfriday"
)

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
  Html() string
  Execute(*HandlerMux) (string, error)
}

type DocumentNode struct {
  NodeList []Node
}

func (d *DocumentNode) Execute(mux *HandlerMux) (string, error) {
  var err error
  var str string

  parts := make([]string, 0, len(d.NodeList))
  for _, node := range d.NodeList {
    str, err = node.Execute(mux)
    if err == nil {
      parts = append(parts, str)
    } else {
      break
    }
  }

  return strings.Join(parts, ""), err
}

func (d *DocumentNode) String() string {
  var parts = make([]string, 0)
  for _, node := range d.NodeList {
    parts = append(parts, node.String())
  }
  return strings.Join(parts, "")
}

func (d *DocumentNode) Html() string {
  var parts = make([]string, 0)
  parts = append(parts, "<div class=\"section\">")
  for _, node := range d.NodeList {
    parts = append(parts, node.Html())
  }
  parts = append(parts, "</div>")
  return strings.Join(parts, "")
}

type BlockTagNode struct {
  Name string
  Subtree Node
}

func (b *BlockTagNode) String() string {
  return magenta + "{{ " + b.Name + " }} " + reset + b.Subtree.String() + magenta + "{{/ " + b.Name + " }}" + reset
}

func (b *BlockTagNode) Html() string {
  return `<span class="block-tag ` + b.Name + `">` + b.Subtree.Html() + `</span>`
}

func (b *BlockTagNode) Execute(mux *HandlerMux) (string, error) {
  return "", nil
}

type TextNode struct {
  Text []byte
}

func (t *TextNode) String() string {
  return string(t.Text)
}

func (t *TextNode) Html() string {
  return `<p>` + string(blackfriday.MarkdownBasic(t.Text)) + `</p>`
}

func (t *TextNode) Execute(mux *HandlerMux) (string, error) {
  return string(t.Text), nil
}

type BraaiTagNode struct {
  Text string
  DotCommands []DotCommandNode
  Arguments []string
  Attributes map[string]string
}

func (b *BraaiTagNode) String() string {
  var parts = make([]string, 2)
  parts[0] = red + "{{ "
  parts[1] = b.Text
  for _, cmd := range b.DotCommands {
    parts = append(parts, cmd.String())
  }

  for _, arg := range b.Arguments {
    parts = append(parts, brightgreen + "\"" + arg + "\", " + reset)
  }

  for k, v := range b.Attributes {
    parts = append(parts, yellow + " " + k + "=" + "\"" + v + "\" " + reset)
  }
  parts = append(parts, red + " }}" + reset)
  return strings.Join(parts, "")
}

func (b *BraaiTagNode) Html() string {
  switch b.Text {
  case "break":
    return `<br />`
  case "attachments":
    return `<div style="color: #f00">This would be an attachment</div>`
  default:
    return `<span class="braai-tag">` + b.Text + `</span>`
  }
}

func (b *BraaiTagNode) Execute(mux *HandlerMux) (string, error) {
  handler := mux.Get(b.Text)
  if handler != nil {
    return handler(b)
  } else {
    return "", fmt.Errorf("Handler not defined for tag: %s", b.Text)
  }
}

type SingleArgumentNode struct {
  Text string
}

func (s *SingleArgumentNode) String() string {
  return blue + "( " + s.Text + " )" + reset
}

func (s *SingleArgumentNode) Html() string {
  return `<span class="single-arg">` + s.Text + `</span>`
}

func (t *SingleArgumentNode) Execute(mux *HandlerMux) (string, error) {
  return t.Text, nil
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

func (d *DotCommandNode) Html() string {
  return `<span class="dot-command">` + d.Text + `</span>`
}

func (d *DotCommandNode) Execute(mux *HandlerMux) (string, error) {
  return "", nil
}

