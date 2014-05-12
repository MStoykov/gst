package gst

/*
#include <gst/gst.h>
*/
import "C"
import "unsafe"

type Clock struct {
	GstObj
}

func (c *Clock) g() *C.GstClock {
	return (*C.GstClock)(unsafe.Pointer(c.Native()))
}

func (c *Clock) AsClock() *Clock {
	return c
}
