package dmp

import "time"

// Handler is the interface for the parsing events
type Handler interface {

	// Object is called on completion of the Object
	// and contains all the properties found.
	Object(obj *Object)

	// Dictionary is called on completion of each dictionary
	//
	// note: because it is called on completion nested children will be
	// returned before their parents.
	Dictionary(dict *Dictionary)

	// Begin is called on the beginning of each section.
	//
	// tag is the section (ex.. "Dictionary", "Object", .. etc).
	// name is the name of the section.
	Begin(tag string, name string)

	// End is called on the ending of each section.
	//
	// tag is the section (ex.. "Dictionary", "Object", .. etc).
	// name is the name of the section.
	//
	// note: end tags in the file typically start with End
	End(tag string, name string)
}

// EmptyHandler is a default implementation of the Handler.
type EmptyHandler struct{}

func (h *EmptyHandler) Object(obj *Object)            {}
func (h *EmptyHandler) Dictionary(dic *Dictionary)    {}
func (h *EmptyHandler) Begin(tag string, name string) {}
func (h *EmptyHandler) End(tag string, name string)   {}

// A Object is the result for Objects in the dmpfile
type Object struct {
	Type       string
	Name       string
	DeviceId   string
	Path       string
	Modified   time.Time
	Properties map[string]string
}

// Table is the structure for the Dictionary entries
type Table struct {
	Header []string
	Rows   [][]string
}

// Dictionary is the result structure for dictionary sections
type Dictionary struct {
	Name   string
	Path   string
	Tables []*Table
}
