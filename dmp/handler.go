package dmp

import "time"

type Handler interface {
	Code(code *Code)
	Object(obj *Object)
	Begin(tag string, name string)
	End(tag string, name string)
}

type EmptyHandler struct{}

func (h *EmptyHandler) Code(code *Code)               {}
func (h *EmptyHandler) Object(obj *Object)            {}
func (h *EmptyHandler) Begin(tag string, name string) {}
func (h *EmptyHandler) End(tag string, name string)   {}

type ForwardingHandler struct{ h Handler }

func (h *ForwardingHandler) Code(code *Code)               { h.h.Code(code) }
func (h *ForwardingHandler) Object(obj *Object)            { h.h.Object(obj) }
func (h *ForwardingHandler) Begin(tag string, name string) { h.h.Begin(tag, name) }
func (h *ForwardingHandler) End(tag string, name string)   { h.h.End(tag, name) }

type Code struct {
	Path     string
	Type     string
	Modified time.Time
	Lines    []string
}

type Object struct {
	Path       string
	Modified   time.Time
	Properties map[string]string
}
