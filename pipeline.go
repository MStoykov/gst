package gst

/*
#include <stdlib.h>
#include <gst/gst.h>
*/
import "C"

import (
	"github.com/conformal/gotk3/glib"
	"unsafe"
)

type Pipeline struct {
	Bin
}

func (p *Pipeline) g() *C.GstPipeline {
	return (*C.GstPipeline)(unsafe.Pointer(p.Native()))
}

func wrapPipeline(obj *glib.Object) *Pipeline {
	return &Pipeline{Bin{Element{GstObj{glib.InitiallyUnowned{obj}}}}}
}

func (p *Pipeline) AsPipeline() *Pipeline {
	return p
}

func NewPipeline(name string) *Pipeline {
	s := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(s))
	c := C.gst_pipeline_new(s)
	if c == nil {
		return nil
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	l := wrapPipeline(obj)
	return l
}
