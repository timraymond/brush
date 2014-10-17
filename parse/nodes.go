package parse

import (
  "fmt"
  "strings"
)

type Node interface{
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

type BlockTagNode struct {
  Name string
  Subtree Node
}

func (b *BlockTagNode) Execute(mux *HandlerMux) (string, error) {
  handler := mux.GetBlock(b.Name)
  if handler != nil {
    return handler(b)
  } else {
    return "", fmt.Errorf("Block Handler not defined for tag: %s", b.Name)
  }
}

type TextNode struct {
  Text []byte
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

func (t *SingleArgumentNode) Execute(mux *HandlerMux) (string, error) {
  return t.Text, nil
}


type DotCommandNode struct {
  Text string
  Argument Node
}

func (d *DotCommandNode) Execute(mux *HandlerMux) (string, error) {
  return "", nil
}

