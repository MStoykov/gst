package gst

/*
#include <stdlib.h>
#include <gst/gst.h>
*/
import "C"

import (
	"github.com/conformal/gotk3/glib"
	"runtime"
	"unsafe"
)

type Bin struct {
	Element
}

func (b *Bin) g() *C.GstBin {
	return (*C.GstBin)(unsafe.Pointer(b.Native()))
}

func (b *Bin) AsBin() *Bin {
	return b
}

func wrapBin(obj *glib.Object) *Bin {
	return &Bin{Element{GstObj{glib.InitiallyUnowned{obj}}}}
}

func NewBin(name string) (*Bin, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gst_bin_new((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	l := wrapBin(obj)
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return l, nil
}

func (b *Bin) Add(els ...*Element) bool {
	for _, e := range els {
		if C.gst_bin_add(b.g(), e.g()) == 0 {
			return false
		}
	}
	return true
}

func (b *Bin) Remove(els ...*Element) bool {
	for _, e := range els {
		if C.gst_bin_remove(b.g(), e.g()) == 0 {
			return false
		}
	}
	return true
}
