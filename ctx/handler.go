package ctx

import (
  "net/http"
)
type Handler struct {
  Context *Context
  Handle func (*Context, http.ResponseWriter, *http.Request)
}
type FinalHandler struct {
  Context *Context
  Payload []byte
  Handle func (*Context, []byte, http.ResponseWriter, *http.Request)
}

func NewHandler(c *Context, f func(c *Context, w http.ResponseWriter, r *http.Request)) (*Handler) {
  return &Handler{c, f}
}
func NewFinalHandler(c *Context, p []byte, f func(c *Context, p[]byte, w http.ResponseWriter, r *http.Request)) (*FinalHandler) {
  return &FinalHandler{c,p,f}
}
func (h Handler) ServeHTTP( w http.ResponseWriter, r *http.Request) {
  h.Handle(h.Context, w, r)
}
func (fh FinalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  fh.Handle(fh.Context, fh.Payload, w, r)
}
