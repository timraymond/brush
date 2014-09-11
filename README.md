Brush
=====

Brush is a compiler for the Braai templating language written in Go.
Braai allows development teams to supply a limited vocabulary of
transformation functions to be used by editorial organizations. This
allows for some flexibility of the presentation of the template while
avoiding having a Turing complete language exposed to CMS users.

Parsing a Braai template is easy. Just include brush and instantiate a
new Braai parser. Next, provide a responder for each identifier that you
would like to support.

```go
package main

import "github.com/reviewed/brush"

const temp = "This is an amazing product: {{product.name}}"

func main() {
  parser, err := brush.New(temp).Parse()
  handlerStack := brush.NewHandleMux()
  handlerStack.Handle("product", func(scope brush.Scope) (string, error) {
    if scope.Env["path"] == ".name" {
      return "Acme Lazer", nil
    } else {
      return nil, fmt.Errorf("I only work in conjunction with a dot handler")
    }
  })
  handlerStack.HandleDot("name", func(scope brush.Scope) brush.Scope {
    scope.Env["path"] = ".name"
    return scope
  })

  result := parser.Exec(handlerStack)
  fmt.Println(result) // This is an amazing product: Acme Lazer
}
```

Each Braai handler has an environment available to it in the form of a
`brush.Scope`. Dot Commands are statements in Braai which transform the
environment in a right-to-left fashion. Environments are scoped
lexically to a braai tag. The command handler is responsible for
outputting the final response, regardless of previous alterations to the
environment. For example, in this template:

```text
Here is an attachment that can be viewed in a modal {{article.attachments(12345).popup}}
```

The DotHandlers for `attachments` and `popup` could set the environment
to look something like this: `attachment_id="12345", mode="popup"`. It
would be the responsibility of the `article` CommandHandler to derive
what those arguments mean, and whether they are valid. Environments can
also be set on the right-most side of a Braai tag, but are overrideable
by dot commands. Consequently, this Braai tag would be equivalent of the
previous example, given the previously stated facts about the dot
commands:

```text
Here is an attachment that can be viewed in a modal {{article attachment_id="12345", mode="popup"}}
```

Block Tags
----------

Tags can also be specified to the Parser to be block-mode tags. These
tags are able to access the global scope only, and receive their
contents as a Braai AST argument. This is useful for creating tags for floating
certain pieces of content, for example. It is the responsibility of the
Block tag for compiling its contents. This also permits block tags to
have a different set of handlers for its contents, as well as altering
the global scope for its subtree.
