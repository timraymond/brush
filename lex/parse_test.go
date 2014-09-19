package lex

import (
  "testing"
  "fmt"
)

type parseTest struct {
  name string
  input string
  ok bool
  result string
}

const (
  noError = true
  hasError = false
)

var parseTests = []parseTest{
  {"text", "Simple *markdown*", noError, `"Simple *markdown*`},
  {"lexical error", "Simple {{?}}", hasError, `Simple {{?}}`},
  {"unterminated braai", "Simple *markdown* {{", hasError, `Simple *markdown* {{`},
  {"simple braai", "The {{product.name}}", noError, `The {{product.name}}`},
  {"simple multi-braai", "The {{product.name}} is called the {{product.name}}", noError, `The {{product.name}} and {{product.name}}`},
  {"simple multi-braai with chained dots", "The {{product.name}} is called the {{product.name.bar}}", noError, `The {{product.name}} and {{product.name}}`},
  {"braai with single arguments", "The {{article.attachments(1235).popup}} is awesome {{product.manufacturer_specs['Color']}}", noError, `The {{product.name}} and {{product.name}}`},
  {"callouts", "You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"multiple callouts", "You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}} some other text {{callout}}I'm in a callout{{/callout}} ending text", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"nested callouts", "You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}} some other text {{callout}}I'm in a callout{{/callout}} ending text", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"attributes", "Here's a photo gallery {{photo_gallery name='blah'}}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"multiple attributes", "Here's a photo gallery {{photo_gallery name='blah', size=\"big\"}}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
}

func TestParse(t *testing.T) {
  for _, test := range parseTests {
    ast, err := New(test.input).Parse()
    if ast != nil {
      fmt.Println(ast)
    }

    if err != nil && test.ok == noError {
      t.Errorf("%s:\n\tUnexpected Parse Error: %s", test.name, err)
    }

    if err == nil && test.ok == hasError {
      t.Errorf("%s:\n\tExpected Parse Error, but saw none", test.name)
    }
  }
}
