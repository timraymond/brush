package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	brush "github.com/timraymond/brush/parse"
)

func Test_Works(t *testing.T) {
	const doc string = "A greeting: {{greeting}}, {{name}}!"
	var err error
	var ast brush.Node
	var result string

	handlers := brush.NewHandlerMux()
	handlers.HandleFunc("greeting", brush.HandlerFunc(func(tag *brush.BraaiTagNode) (string, error) {
		return "Hello there", nil
	}))
	handlers.HandleFunc("name", brush.HandlerFunc(func(tag *brush.BraaiTagNode) (string, error) {
		return "tim", nil
	}))

	ast, err = brush.New("exectest", doc, []string{}).Parse()
	if assert.NoError(t, err) {
		result, err = ast.Execute(handlers)
		if assert.NoError(t, err) {
			assert.Equal(t, "A greeting: Hello there, tim!", result)
		}
	}
}

func Test_DiesIfNoHandler(t *testing.T) {
	const doc string = "A greeting: {{greeting}}, {{name}}!"
	var err error
	var ast brush.Node

	handlers := brush.NewHandlerMux()
	handlers.HandleFunc("name", brush.HandlerFunc(func(tag *brush.BraaiTagNode) (string, error) {
		return "tim", nil
	}))

	ast, err = brush.New("exectest", doc, []string{}).Parse()
	if assert.NoError(t, err) {
		_, err = ast.Execute(handlers)
		if assert.Error(t, err) {
			assert.Equal(t, "exectest:1:14: Exec error - Handler not defined for tag: [greeting]", err.Error())
		}
	}
}

type ProductHandler struct {
	ProductName string
}

func (p *ProductHandler) Name(modelNum string) (string, error) {
	return p.ProductName + " " + modelNum, nil
}

func (p *ProductHandler) Adjective() (string, error) {
	return "greatest", nil
}

func Test_WithDotCommands(t *testing.T) {
	const doc string = "The {{product.name['9000']}} is the {{product.adjective}} camera"
	var err error
	var ast brush.Node
	var result string

	handlers := brush.NewHandlerMux()
	handlers.Handle("product", &ProductHandler{"Canon Foo"})

	ast, err = brush.New("exectest", doc, []string{}).Parse()
	if assert.NoError(t, err) {
		result, err = ast.Execute(handlers)
		if assert.NoError(t, err) {
			assert.Equal(t, "The Canon Foo 9000 is the greatest camera", result)
		}
	}
}

func Test_WithDotCommandsShouldExplode(t *testing.T) {
	const doc string = "The {{product.name['9000']}} is the {{product.verb}} camera"
	var err error
	var ast brush.Node

	handlers := brush.NewHandlerMux()
	handlers.Handle("product", &ProductHandler{"Canon Foo"})

	ast, err = brush.New("exectest", doc, []string{}).Parse()
	if assert.NoError(t, err) {
		_, err = ast.Execute(handlers)
		if assert.Error(t, err) {
			assert.Equal(t, "Undefined method `Verb` for product handler", err.Error())
		}
	}
}

func Test_BlockHandlers(t *testing.T) {
	const doc string = "Here's some text with {{bold}}emphasis{{/bold}}"

	handlers := brush.NewHandlerMux()
	handlers.HandleBlockFunc("bold", func(tag *brush.BlockTagNode) (string, error) {
		subtree, err := tag.Subtree.Execute(handlers)
		if err != nil {
			return "", err
		} else {
			return "<bold>" + subtree + "</bold>", nil
		}
	})

	ast, err := brush.New("exectest", doc, handlers.BlockHandlers()).Parse()
	if assert.NoError(t, err) {
		result, err := ast.Execute(handlers)
		if assert.NoError(t, err) {
			assert.Equal(t, "Here's some text with <bold>emphasis</bold>", result)
		}
	}
}

func Test_Default_Handlers(t *testing.T) {
	const doc string = "Tag 1: {{foo}}, Tag 2: {{bar}}"

	tags := make([]string, 0, 2)
	handlers := brush.NewHandlerMux()
	handlers.DefaultHandler(func(tag *brush.BraaiTagNode) (string, error) {
		tags = append(tags, tag.Text)
		return "{{ " + tag.Text + "}}", nil
	})
	ast, err := brush.New("default", doc, handlers.BlockHandlers()).Parse()
	if assert.NoError(t, err) {
		_, err := ast.Execute(handlers)
		if assert.NoError(t, err) {
			assert.Equal(t, tags, []string{"foo", "bar"})
		}
	}
}

type TestVisitor struct {
	Ids []string
}

func (tv *TestVisitor) AcceptTag(tag *brush.BraaiTagNode) {
	tv.Ids = append(tv.Ids, tag.Attributes["id"])
}

func (tv *TestVisitor) AcceptBlockTag(tag *brush.BlockTagNode) {
	// NOP
}

func (tv *TestVisitor) AcceptTextNode(tag *brush.TextNode) {
	// NOP
}

func (tv *TestVisitor) String() string {
	return "Here are my ids: " + strings.Join(tv.Ids, ", ")
}

func Test_Visitors(t *testing.T) {
	const doc string = `Here's a bunch of tags: {{ foo id="12345" }} {{ foo id="45678" }}`

	expected := []string{
		"12345",
		"45678",
	}

	tv := &TestVisitor{make([]string, 0, 2)}
	handlers := brush.NewHandlerMux()
	ast, err := brush.New("default", doc, handlers.BlockHandlers()).Parse()

	if err != nil {
		t.Errorf("Visitor Test: encountered error: %s", err.Error())
	}

	ast.Visit(tv)
	for i, elem := range expected {
		if tv.Ids[i] != elem {
			t.Errorf("Visitor Test: Expected %s to equal %s", tv.Ids[i], elem)
		}
	}
}

func ExampleCompositeVisitor() {
	document := `
  This is a document with some tags that have ids:
  {{ foo id="4815162342" }}
  {{ foo id="8675309" }}
  `
	testVisitor := &TestVisitor{make([]string, 0, 2)}
	cv := brush.NewCompositeVisitor(testVisitor)
	handlers := brush.NewHandlerMux()
	ast, _ := brush.New("default", document, handlers.BlockHandlers()).Parse()

	ast.Visit(cv)

	fmt.Println(testVisitor)
	// Output: Here are my ids: 4815162342, 8675309
}
