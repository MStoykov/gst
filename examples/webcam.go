package main

import (
	"github.com/tobert/glib"
	"github.com/tobert/gst"
)

// USB2 can't handle a raw 1080p stream, so cameras that offer it
// also offer h264 encoded video, this is how you get it
// tested with a Logitech C920 HD
var CAMERA_FORMAT = "video/x-h264, width=1920, height=1080, framerate=24/1"

func main() {
	pl := gst.NewPipeline("webcam")

	// grab video from /dev/video0, assumed to be a webcam
	vsrc := gst.ElementFactoryMake("v4l2src", "webcam")
	vsrc.SetProperty("device", "/dev/video0")
	pl.Add(vsrc)

	vcap := gst.CapsFromString(CAMERA_FORMAT)
	defer vcap.Unref()

	// decode the h.264 stream to raw
	vdecoder := gst.ElementFactoryMake("avdec_h264", "video_decoder")
	pl.Add(vdecoder)
	vsrc.LinkFiltered(vdecoder, vcap)

	// display using the clutter sink
	sink := gst.ElementFactoryMake("autocluttersink", "display")
	pl.Add(sink)
	vdecoder.Link(sink)

	pl.SetState(gst.STATE_PLAYING)
	glib.NewMainLoop(nil).Run()
}

// vim: ts=4 sw=4 noet tw=120 softtabstop=4
