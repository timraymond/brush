package parse

import (
  "fmt"
  "reflect"
  "strings"
)

type HandlerMux struct {
  funcs map[string]HandlerFunc
}

type HandlerFunc func(*BraaiTagNode) (string, error)

func (h *HandlerMux) HandleFunc(ident string, f HandlerFunc) {
  h.funcs[ident] = f
}

// Defines a HandlerFunc which introspects the passed in handler, invoking
// methods matching dot command nodes.
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

func (h *HandlerMux) Get(name string) HandlerFunc {
  return h.funcs[name]
}

func NewHandlerMux() *HandlerMux {
  mux := &HandlerMux{}
  mux.funcs = make(map[string]HandlerFunc)
  return mux
}
