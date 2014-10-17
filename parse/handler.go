package parse

import (
  "fmt"
  "reflect"
  "strings"
)

// A HandlerMux is a collection of the user-specified functions used for
// transforming particular Braai tags into strings.
type HandlerMux struct {
  funcs map[string]HandlerFunc
  blockFuncs map[string]BlockHandlerFunc
}

// A HandlerFunc is a function which receives the raw BraaiTagNode, and is
// expected to return the finished content
type HandlerFunc func(*BraaiTagNode) (string, error)

// A BlockHandlerFunc receives a BlockTagNode, and is expected to return the
// finished content. It is considered low-level, in that it must invoke the
// Execute() method on the BlockTagNode's Subtree for its children to be
// rendered properly
type BlockHandlerFunc func(*BlockTagNode) (string, error)

// HandleFunc registers a HandlerFunc with this HandlerMux
func (h *HandlerMux) HandleFunc(ident string, f HandlerFunc) {
  h.funcs[ident] = f
}

// HandleBlockFunc registers a BlockHandlerFunc with this HandlerMux
func (h *HandlerMux) HandleBlockFunc(ident string, f func(*BlockTagNode) (string, error)) {
  h.blockFuncs[ident] = BlockHandlerFunc(f)
}

// HandleFuncWrap takes a slice strings, naming all types of BraaiTag found in
// the Subtree of this Block handler, it is expected to return two strings for
// the prefix and suffix of the rendered subtree's content, and an error,
// should one occur
func (h *HandlerMux) HandleFuncWrap(ident string, wrap func([]string) (string, string, error)) {
  // TODO
}

// BlockHandlers returns a []string of all the identifiers which have block
// handlers associated with them. This is intended to make it easy to specify
// which identifiers should be considered block braai tags.
func (h *HandlerMux) BlockHandlers() (handlers []string) {
  for name, _ := range h.blockFuncs {
    handlers = append(handlers, name)
  }
  return handlers
}

// Handle defines a HandlerFunc which introspects the passed in handler,
// invoking methods matching dot command nodes.
func (h *HandlerMux) Handle(ident string, handler interface{}) {
  h.funcs[ident] = HandlerFunc(func(b *BraaiTagNode) (string, error) {
    for _, cmd := range b.DotCommands {
      method := reflect.ValueOf(handler).MethodByName(strings.Title(cmd.Text))
      if method.IsValid() == false {
        return "", fmt.Errorf("Undefined method `%s` for %s handler", strings.Title(cmd.Text), ident)
      }
      if cmd.Argument != nil {
        argument, _ := cmd.Argument.Execute(h)
        return method.Interface().(func(string)(string, error))(argument)
      } else {
        return method.Interface().(func()(string, error))()
      }
    }
    return "", nil
  })
}

// Get returns a previously defined HandlerFunc using either Handle or
// HandleFunc
func (h *HandlerMux) Get(name string) HandlerFunc {
  return h.funcs[name]
}

// GetBlock returns a previously defined BlockHandlerFunc using HandleBlockFunc
func (h *HandlerMux) GetBlock(name string) BlockHandlerFunc {
  return h.blockFuncs[name]
}

// NewHandlerMux returns a new HandlerMux with internal maps pre-initialized.
// All HandlerMuxes should be created this way to ensure future initialization
// logic is handled
func NewHandlerMux() *HandlerMux {
  mux := &HandlerMux{}
  mux.funcs = make(map[string]HandlerFunc)
  mux.blockFuncs = make(map[string]BlockHandlerFunc)
  return mux
}
