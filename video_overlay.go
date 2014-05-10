package gst

/*
#include <gst/video/videooverlay.h>

GstVideoOverlay* _gst_video_overlay_cast(GstObject* o) {
	return GST_VIDEO_OVERLAY(o);
}
#cgo pkg-config: gstreamer-video-1.0
*/
import "C"

import (
	"github.com/ziutek/glib"
)

type VideoOverlay C.GstVideoOverlay

func (x *VideoOverlay) g() *C.GstVideoOverlay {
	return (*C.GstVideoOverlay)(x)
}

func (x *VideoOverlay) Type() glib.Type {
	return glib.TypeFromName("GstVideoOverlay")
}

func (x *VideoOverlay) SetVideoWindowHandle(id uint) {
	C.gst_video_overlay_set_window_handle(x.g(), C.guintptr(id))
}

func (x *VideoOverlay) GotVideoWindowHandle(id uint) {
	C.gst_video_overlay_got_window_handle(x.g(), C.guintptr(id))
}

func (x *VideoOverlay) PrepareWindowId() {
	C.gst_video_overlay_prepare_window_handle(x.g())
}

func (x *VideoOverlay) Expose() {
	C.gst_video_overlay_expose(x.g())
}

func (x *VideoOverlay) HandleEvents(handle_events bool) {
	var he C.gboolean
	if handle_events {
		he = 1
	}
	C.gst_video_overlay_handle_events(x.g(), he)
}

func (o *VideoOverlay) SetRenderRectangle(x, y, width, height int) bool {
	return C.gst_video_overlay_set_render_rectangle(o.g(), C.gint(x), C.gint(y),
		C.gint(width), C.gint(height)) != 0
}

func VideoOverlayCast(o *GstObj) *VideoOverlay {
	return (*VideoOverlay)(C._gst_video_overlay_cast(o.g()))
}
