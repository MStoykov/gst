// Bindings for GStreamer API
package gst

/*
#include <stdlib.h>
#include <gst/gst.h>

char** _gst_init(int* argc, char** argv) {
	gst_init(argc, &argv);
	return argv;
}

typedef struct {
	const char *name;
	const GValue *val;
} Field;

typedef struct {
	Field* tab;
	int    n;
} Fields;

gboolean _parse_field(GQuark id, const GValue* val, gpointer data) {
	Fields *f = (Fields*)(data);
	f->tab[f->n].name = g_quark_to_string(id);
	f->tab[f->n].val = val;
	++f->n;
	return TRUE;
}

Fields _parse_struct(GstStructure *s) {
	int n = gst_structure_n_fields(s);
	Fields f = { malloc(n * sizeof(Field)), 0 };
	gst_structure_foreach(s, _parse_field, (gpointer)(&f));
	return f;
}

#cgo pkg-config: gstreamer-1.0
*/
import "C"

import (
	"errors"
	_ "fmt"
	"github.com/conformal/gotk3/glib"
	"os"
	"unsafe"
)

func v2g(v *glib.Value) *C.GValue {
	return (*C.GValue)(unsafe.Pointer(v))
}

func g2v(v *C.GValue) *glib.Value {
	return (*glib.Value)(unsafe.Pointer(v))
}

type Fourcc C.guint32

var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

func (f Fourcc) String() string {
	buf := make([]byte, 4)
	buf[0] = byte(f)
	buf[1] = byte(f >> 8)
	buf[2] = byte(f >> 16)
	buf[3] = byte(f >> 32)
	return string(buf)
}

func MakeFourcc(a, b, c, d byte) Fourcc {
	return Fourcc(uint32(a) | uint32(b)<<8 | uint32(c)<<16 | uint32(d)<<24)
}

func StrFourcc(s string) Fourcc {
	if len(s) != 4 {
		panic("Fourcc string length != 4")
	}
	return MakeFourcc(s[0], s[1], s[2], s[3])
}
func init() {
	alen := C.int(len(os.Args))
	argv := make([]*C.char, alen)
	for i, s := range os.Args {
		argv[i] = C.CString(s)
	}
	ret := C._gst_init(&alen, &argv[0])
	argv = (*[1 << 16]*C.char)(unsafe.Pointer(ret))[:alen]
	os.Args = make([]string, alen)
	for i, s := range argv {
		os.Args[i] = C.GoString(s)
	}
	tm := []glib.TypeMarshaler{
		{glib.Type(C.gst_bus_get_type()), marshalBus},
		{glib.Type(C.gst_message_get_type()), marshalMessage},
	}

	glib.RegisterGValueMarshalers(tm)
}

func marshalBus(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return wrapBus(obj), nil
}

func marshalMessage(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return (*Message)(unsafe.Pointer(c)), nil
}

func makeGstStructure(name string, fields map[string]interface{}) *C.GstStructure {
	nm := (*C.gchar)(C.CString(name))
	s := C.gst_structure_new_empty(nm)
	C.free(unsafe.Pointer(nm))
	for k, v := range fields {
		n := (*C.gchar)(C.CString(k))
		value, err := glib.GValue(v)
		if err != nil {
			panic(err)
		}
		C.gst_structure_take_value(s, n, v2g(value))
		C.free(unsafe.Pointer(n))
	}
	return s
}

func parseGstStructure(s *C.GstStructure) (name string, fields map[string]interface{}) {
	name = C.GoString((*C.char)(C.gst_structure_get_name(s)))
	ps := C._parse_struct(s)
	n := (int)(ps.n)
	tab := (*[1 << 16]C.Field)(unsafe.Pointer(ps.tab))[:n]
	fields = make(map[string]interface{})
	for _, f := range tab {
		value, err := g2v(f.val).GoValue()
		if err != nil {
			panic(err)
		}
		fields[C.GoString(f.name)] = value
	}
	return
}

var CLOCK_TIME_NONE = int64(C.GST_CLOCK_TIME_NONE)

var GST_VERSION_MINOR = int64(C.GST_VERSION_MINOR)
var GST_VERSION_MICRO = int64(C.GST_VERSION_MICRO)
var GST_VERSION_NANO = int64(C.GST_VERSION_NANO)

func Version() (int, int, int, int) {
	var cmajor, cminor, cmicro, cnano C.guint
	C.gst_version(&cmajor, &cminor, &cmicro, &cnano)
	return int(cmajor), int(cminor), int(cmicro), int(cnano)
}

func VersionString() string {
	return C.GoString((*C.char)(C.gst_version_string()))
}
