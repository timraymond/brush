package parse_test

import (
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

  ast, err = brush.New(doc, []string{}).Parse()
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

  ast, err = brush.New(doc, []string{}).Parse()
  if assert.NoError(t, err) {
    _, err = ast.Execute(handlers)
    if assert.Error(t, err) {
      assert.Equal(t, "Handler not defined for tag: greeting", err.Error())
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

  ast, err = brush.New(doc, []string{}).Parse()
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

  ast, err = brush.New(doc, []string{}).Parse()
  if assert.NoError(t, err) {
    _, err = ast.Execute(handlers)
    if assert.Error(t, err) {
      assert.Equal(t, "Undefined method `Verb` for product handler", err.Error())
    }
  }
}
