package parse

import (
	"fmt"
	"strings"
)

// A Node represents any struct which is able to traverse itself (and any of
// its children, if applicable) with the passed in HandlerMux. It signifies a
// Node in a Braai AST.
type Node interface {
	Execute(*HandlerMux) (string, error)
}

// A DocumentNode represents a complete Braai document. There are no
// restrictions as to where these can appear in the document to support things
// such as including other documents and also representing the Subtrees of a
// BlockTagNode
type DocumentNode struct {
	NodeList []Node
}

// Execute implements the Node interface, and invokes Execute on every member
// of the  DocumentNode's NodeList, assembling each fragment into a whole to be
// returned to the caller
func (d *DocumentNode) Execute(mux *HandlerMux) (str string, err error) {
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

// A BlockTagNode represents a block-form BraaiTag, such as foo in this example:
//   {{foo}}Content {{bar(1234)}}{{/foo}}
// All content within the Block is provided as the Subtree Node.
type BlockTagNode struct {
	Name    string
	Subtree Node
}

// Execute searches for a registered block tag handler within the HandlerMux,
// invoking it if present. It is expected that this handler will compile the
// Subtree, but it is not required
func (b *BlockTagNode) Execute(mux *HandlerMux) (string, error) {
	handler := mux.GetBlock(b.Name)
	if handler != nil {
		return handler(b)
	} else {
		return "", fmt.Errorf("Block Handler not defined for tag: %s", b.Name)
	}
}

// A TextNode represents text devoid of any Braai tags. These are left
// unmodified by handlers.
type TextNode struct {
	Text []byte
}

// Execute passes the TextNode's Text  back to the caller
func (t *TextNode) Execute(mux *HandlerMux) (string, error) {
	return string(t.Text), nil
}

// A BraaiTagNode represents a non-block Braai tag. All DotCommands, Arguments,
// and Attributes for the tag are also stored here
type BraaiTagNode struct {
	Text        string
	DotCommands []DotCommandNode
	Arguments   []string
	Attributes  map[string]string
	Pos         string
}

// Execute searches for a HandlerFunc for this BraaiTag and invokes it if
// found.
func (b *BraaiTagNode) Execute(mux *HandlerMux) (string, error) {
	handler := mux.Get(b.Text)
	if handler != nil {
		return handler(b)
	} else {
		handler = mux.GetDefaultHandler()
		if handler != nil {
			return handler(b)
		} else {
			return "", b.Errorf("Handler not defined for tag: %s", b.Text)
		}
	}
}

func (b *BraaiTagNode) Errorf(format string, args ...interface{}) error {
	format = b.Pos + "Exec error - " + format
	return fmt.Errorf(format, args)
}

// A SingleArgumentNode represents an argument appearing in parentheses
// following a top level command or one following a dot command.
type SingleArgumentNode struct {
	Text string
}

// Execute returns the text of the argument. It is assumed that a
// BraaiTagHandler will manipulate this directly in the BraaiTag
func (t *SingleArgumentNode) Execute(mux *HandlerMux) (string, error) {
	return t.Text, nil
}

// A DotCommandNode represents a command appearing after the primary command in
// a braai tag. For example, in the following Braai tag:
//   {{user.first_name}}
// `first_name` would be a dot command. These can be processed automatically
// using the Handle() method of the HandlerMux
type DotCommandNode struct {
	Text     string
	Argument Node
}

// Execute is effectively a no-op here, since it is assumed to be handled by
// the BraaiTag handler. This is only here to implement the Node interface
func (d *DotCommandNode) Execute(mux *HandlerMux) (string, error) {
	return "", nil
}
