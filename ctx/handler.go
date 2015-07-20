package ctx

import (
  "net/http"
)
type Handler struct {
  Context *Context
  Handle func (*Context, http.ResponseWriter, *http.Request)
}
func (ah Handler) ServeHTTP( w http.ResponseWriter, r *http.Request) {
  ah.Handle(ah.Context, w, r)
}
