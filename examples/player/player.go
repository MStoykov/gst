package main

import (
	"log"
	"time"

	"github.com/MStoykov/gdk_x11"
	"github.com/MStoykov/gst"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/gtk"
)

type FileChooserWindow struct {
	*gtk.Window
	FileChooserWidget *gtk.FileChooserWidget
}

func newFileChooserWindow() *FileChooserWindow {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Simple Example")

	widget, err := gtk.FileChooserWidgetNew(gtk.FILE_CHOOSER_ACTION_OPEN)
	if err != nil {
		log.Fatal("Unable to create filechooser:", err)
	}

	win.Add(widget)

	return &FileChooserWindow{win, widget}
}

type Player struct {
	pipe       *gst.Element
	bus        *gst.Bus
	window     *gtk.Window
	movie_area *gtk.DrawingArea
	file_path  string
	xid        uint
}

func (p *Player) onPlayClicked(b *gtk.Button) {
	if p.file_path != "" {
		err := p.pipe.SetProperty("uri", "file://"+p.file_path)
		if err != nil {
			panic(err)
		}
		p.pipe.SetState(gst.STATE_PLAYING)
	}
}

func (p *Player) onPauseClicked(b *gtk.Button) {
	state, _, _ := p.pipe.GetState(gst.CLOCK_TIME_NONE)
	if state == gst.STATE_PLAYING {
		p.pipe.SetState(gst.STATE_PAUSED)
	}
}

func (p *Player) onStopClicked(b *gtk.Button) {
	p.pipe.SetState(gst.STATE_NULL)
}

func (p *Player) onFileSelected(widget *gtk.FileChooserWidget) {
	p.file_path = widget.FileChooser.GetFilename()
}

func (p *Player) onMessage(bus *gst.Bus, msg *gst.Message) {
	switch msg.GetType() {
	case gst.MESSAGE_EOS:
		p.pipe.SetState(gst.STATE_NULL)
	case gst.MESSAGE_ERROR:
		p.pipe.SetState(gst.STATE_NULL)
		err, debug := msg.ParseError()
		log.Printf("Error: %s (debug: %s)\n", err, debug)
	}
}

func (p *Player) onSyncMessage(bus *gst.Bus, msg *gst.Message) {
	name, _ := msg.GetStructure()
	if name != "prepare-window-handle" {
		return
	}
	img_sink := msg.GetSrc()
	xov := gst.VideoOverlayCast(img_sink)
	if p.xid != 0 && xov != nil {
		img_sink.Set("force-aspect-ratio", true)
		xov.SetVideoWindowHandle(p.xid)
	} else {
		log.Println("Error: xid =", p.xid, "xov =", xov)
	}
}

func (p *Player) onVideoWidgetRealize(w *gtk.DrawingArea) {
	window, err := w.GetWindow()
	if err != nil {
		log.Fatal("Error on getting movie_area Window", err)
	}
	p.xid = gdk_x11.GetGDKX11WindowId(window)
}

func NewPlayer() *Player {
	p := new(Player)

	var err error
	p.window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		panic(err)
	}
	p.window.SetTitle("Player")
	p.window.Connect("destroy", gtk.MainQuit)

	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		panic(err)
	}
	p.window.Add(vbox)
	hbox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		panic(err)
	}
	vbox.PackStart(hbox, false, false, 0)

	fcb := NewFileChooserButton(
		"Choose media file",
		gtk.FILE_CHOOSER_ACTION_OPEN,
	)
	fcb.window.FileChooserWidget.Connect("file-activated", p.onFileSelected)
	hbox.Add(fcb)

	button, err := gtk.ButtonNewFromIconName("gtk-media-play", gtk.ICON_SIZE_BUTTON)
	if err != nil {
		panic(err)
	}
	button.Connect("clicked", p.onPlayClicked)
	hbox.PackStart(button, false, false, 0)

	button, err = gtk.ButtonNewFromIconName("gtk-media-pause", gtk.ICON_SIZE_BUTTON)
	if err != nil {
		panic(err)
	}
	button.Connect("clicked", p.onPauseClicked)
	hbox.PackStart(button, false, false, 0)

	button, err = gtk.ButtonNewFromIconName("gtk-media-stop", gtk.ICON_SIZE_BUTTON)
	if err != nil {
		panic(err)
	}
	button.Connect("clicked", p.onStopClicked)
	hbox.PackStart(button, false, false, 0)

	p.movie_area, _ = gtk.DrawingAreaNew()
	p.movie_area.Connect("realize", p.onVideoWidgetRealize)
	p.movie_area.SetDoubleBuffered(false)
	p.movie_area.SetSizeRequest(640, 360)
	vbox.Add(p.movie_area)

	p.window.ShowAll()
	p.pipe = gst.ElementFactoryMake("playbin", "mine")
	p.bus = p.pipe.GetBus()
	p.bus.AddSignalWatch()

	p.bus.Connect("message", p.onMessage)
	p.bus.EnableSyncMessageEmission()
	p.bus.Connect("sync-message", p.onSyncMessage)
	p.window.Connect("key-press-event", p.onKeyPress)

	return p
}

func (p *Player) jumpBy(by int64) {
	var d = new(int64)
	b := p.pipe.QueryPosition(gst.FORMAT_TIME, d)
	if b == false {
		log.Printf("Couldn't get current position while trying to jump")
		return
	}
	*d = *d + by
	b = p.pipe.SeekSimple(gst.FORMAT_TIME, gst.SEEK_FLAG_FLUSH|gst.SEEK_FLAG_KEY_UNIT, *d)
	if b == false {
		log.Printf("Couldn't get change position while trying to jump")
	}
}

func (p *Player) jumpForward() {
	p.jumpBy(int64(time.Second))
}

func (p *Player) jumpBackward() {
	p.jumpBy(-int64(time.Second))
}

func (p *Player) onKeyPress(w *gtk.Window, event *gdk.Event) bool {
	if event.GetType() != gdk.GDK_KEY_PRESS {
		panic("wrong type of event in onKeyPress handler")
	}
	eventKey := gdk.EventKeyFromEvent(event)
	keyVal, err := eventKey.GetKeyVal()
	if err != nil {
		panic(err)
	}
	switch keyVal {
	case gdk.GDK_KEY_Left:
		log.Println("left")
		p.jumpBackward()
	case gdk.GDK_KEY_Right:
		log.Println("right")
		p.jumpForward()
	case gdk.GDK_KEY_a:
		log.Println("a")
		var d = new(int64)
		b := p.pipe.QueryPosition(gst.FORMAT_TIME, d)
		if b == true {
			log.Printf("current position: %s\n", b, time.Duration(*d))
		}
	}

	log.Println("key pressed")
	return false
}

type FileChooseButton struct {
	*gtk.Button
	window *FileChooserWindow
}

func NewFileChooserButton(label string, action gtk.FileChooserAction) *FileChooseButton {
	button, err := gtk.ButtonNewWithLabel(label)
	if err != nil {
		panic(err)
	}
	window := newFileChooserWindow()
	button.Connect("clicked", func() {
		window.ShowAll()
	})
	return &FileChooseButton{button, window}
}

func (p *Player) Run() {
	gtk.Main()
}

func main() {
	gtk.Init(nil)
	player := NewPlayer()
	player.Run()
}
