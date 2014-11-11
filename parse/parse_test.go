package parse

import (
  "testing"
)

type parseTest struct {
  name string
  input string
  ok bool
  errorMsg string
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
  {"argument list", "Egads, it's a gallery with an arglist {{ photo_gallery \"Ashtray\", \"Garbage Can\", \"Doorknob\" }}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"a menagerie of braai", "And all together now! {{callout}}{{ photo_gallery \"Ashtray\", \"Garbage Can\", \"Doorknob\" size=\"big\" }}{{/callout}}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"a menagerie of braai, and dot commands", "And all together now! {{callout}}{{ article.attachments(1234) \"Ashtray\", \"Garbage Can\", \"Doorknob\" size=\"big\" }}{{/callout}}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"a menagerie of braai, and dot commands", "And all together now! {{callout}}{{ article.attachments[\"Upper Deck\"] \"Ashtray\", \"Garbage Can\", \"Doorknob\" size=\"big\" }}{{/callout}}", noError, `You've got an attachment in my callout! {{callout}}{{article.attachments(1235).popup}}{{/callout}}`},
  {"float right", "This should be floated right: {{float_right}}{{ attachments(346360).popup }}{{/float_right}}", noError, "This should be floated right: {{float_right}}{{ article.attachments(12345).popup }}{{/float_right}}"},
}

var errorTests = []parseTest{
  // Ensure line numbers work
  {"unterminated", "Foo {{photo_gallery}", hasError, `unterminated:1:19: Lexical Error - Malformed end of Braai tag, should be }}`},
  {"invalidchar", "Foo\n\n{{foo?}}", hasError, `invalidchar:3:6: Lexical Error - Unexpected character U+003F '?'`},
}

func TestParse(t *testing.T) {
  for _, test := range parseTests {
    _, err := New(test.name, test.input, []string{"callout", "float_right"}).Parse()

    if err != nil && test.ok == noError {
      t.Errorf("%s:\n\tUnexpected Parse Error: %s", test.name, err)
    }

    if err == nil && test.ok == hasError {
      t.Errorf("%s:\n\tExpected Parse Error, but saw none", test.name)
    }
  }
}

func TestParseErrors(t *testing.T) {
  for _, test := range errorTests {
    _, err := New(test.name, test.input, []string{}).Parse()

    if err != nil && test.errorMsg != err.Error() {
      t.Errorf("%s\n\tError message mismatch. Expected: \"%s\", Actual: \"%s\"", test.name, test.errorMsg, err.Error())
    }
    
    if err == nil {
      t.Errorf("%s\n\tNo parse errors, place this test in parseTests", test.name)
    }
  }
}

