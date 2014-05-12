package main

import (
	"fmt"
	"github.com/MStoykov/gdk_x11"
	"github.com/MStoykov/gst"
	"github.com/conformal/gotk3/gtk"
	"github.com/davecgh/go-spew/spew"
	"log"
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
	log.Println("play-clicked")
	log.Println("file", p.file_path)
	if p.file_path != "" {
		spew.Dump(p.pipe)
		err := p.pipe.SetProperty("uri", "file://"+p.file_path)
		if err != nil {
			panic(err)
		}
		prop, err := p.pipe.GetProperty("uri")
		log.Printf("prop: %s, err :%s", prop, err)
		p.pipe.SetState(gst.STATE_PLAYING)

		prop, err = p.pipe.GetProperty("current-uri")
		log.Printf("prop: %s, err :%s", prop, err)
		prop, err = p.pipe.GetProperty("current-urisds")
		log.Printf("prop: %s, err :%s", prop, err)
		spew.Dump(p.pipe)
	}

	log.Println("play-clicked-end")
}

func (p *Player) onPauseClicked(b *gtk.Button) {
	log.Println("pause-clicked")
	state, _, _ := p.pipe.GetState(gst.CLOCK_TIME_NONE)
	if state == gst.STATE_PLAYING {
		p.pipe.SetState(gst.STATE_PAUSED)
	}
}

func (p *Player) onStopClicked(b *gtk.Button) {

	log.Println("stop-clicked")
	p.pipe.SetState(gst.STATE_NULL)
}

func (p *Player) onFileSelected(widget *gtk.FileChooserWidget) {
	p.file_path = widget.FileChooser.GetFilename()
	log.Println("file-selected", p.file_path)
}

func (p *Player) onMessage(bus *gst.Bus, msg *gst.Message) {
	//name, err := msg.GetStructure()
	//log.Printf("name: %s ,  err  %s\n", name, err)
	switch msg.GetType() {
	case gst.MESSAGE_EOS:
		p.pipe.SetState(gst.STATE_NULL)
	case gst.MESSAGE_ERROR:
		p.pipe.SetState(gst.STATE_NULL)
		err, debug := msg.ParseError()
		fmt.Printf("Error: %s (debug: %s)\n", err, debug)
	}
}

func (p *Player) onSyncMessage(bus *gst.Bus, msg *gst.Message) {
	name, _ := msg.GetStructure()
	/*
		log.Println("!!!!!!!!!!!!!!OnSyncMessage")
		log.Printf("name: %s\n", name)
		for name, value := range str {
			log.Printf("---> %s: %s", name, value)
		}
	*/
	if name != "prepare-window-handle" {
		return
	}
	img_sink := msg.GetSrc()
	xov := gst.VideoOverlayCast(img_sink)
	if p.xid != 0 && xov != nil {
		img_sink.Set("force-aspect-ratio", true)
		xov.SetVideoWindowHandle(p.xid)
	} else {
		fmt.Println("Error: xid =", p.xid, "xov =", xov)
	}
}

func (p *Player) onVideoWidgetRealize(w *gtk.DrawingArea) {
	fmt.Println("realized")
	window, err := w.GetWindow()
	if err != nil {
		log.Fatal("Error on getting movie_area Window", err)
	}
	p.xid = gdk_x11.GetGDKX11WindowId(window)
	fmt.Printf("window-id:%d ", p.xid)
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
	//p.movie_area.SetDoubleBuffered(false)
	p.movie_area.SetSizeRequest(640, 360)
	vbox.Add(p.movie_area)

	p.window.ShowAll()
	//p.window.Realize()
	p.pipe = gst.ElementFactoryMake("playbin", "mine")
	p.bus = p.pipe.GetBus()
	p.bus.AddSignalWatch()

	log.Printf("p.bus: %#v", p.bus)
	p.bus.Connect("message", p.onMessage)
	p.bus.EnableSyncMessageEmission()
	p.bus.Connect("sync-message", p.onSyncMessage)

	return p
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
	NewPlayer().Run()
}
