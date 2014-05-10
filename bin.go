package gst

/*
#include <stdlib.h>
#include <gst/gst.h>
*/
import "C"

import (
	"github.com/ziutek/glib"
	"unsafe"
)

type Bin struct {
	Element
}

func (b *Bin) g() *C.GstBin {
	return (*C.GstBin)(b.GetPtr())
}

func (b *Bin) AsBin() *Bin {
	return b
}

func NewBin(name string) *Bin {
	s := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(s))
	b := new(Bin)
	b.SetPtr(glib.Pointer(C.gst_bin_new(s)))
	return b
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
