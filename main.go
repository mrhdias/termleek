//
// Copyright (C) 2023 Henrique Dias <mrhdias@gmail.com>
// MIT License
//
// References:
// https://developer.gnome.org/vte/unstable/VteTerminal.html
// https://www.spinics.net/lists/gtk/msg21846.html
// https://ini.unknwon.io/docs/intro/getting_started
// https://cpp.hotexamples.com/pt/examples/-/-/pango_font_description_free/cpp-pango_font_description_free-function-examples.html
// https://github.com/nlamirault/mert/blob/master/vte3/vte3.go
// https://github.com/orhun/kermit
//

package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/mrhdias/vte"
	vteGtk "github.com/mrhdias/vte/vte.gtk3"
	"gopkg.in/ini.v1"
)

const (
	minWidth  = 340
	minHeight = 185
)

type App struct {
	Config struct {
		Background struct {
			Source              string
			PreserveAspectRatio bool
		}
		Terminal struct {
			MinWidth  int
			MinHeight int
			Opacity   float64
			Icon      string
			Font      string
		}
	}
	Width           int
	Height          int
	BackgroundImage *gtk.Image
	Window          *gtk.Window
}

func (app App) resizeImage(widget *gtk.Window, data gdk.Pixbuf) *gdk.Pixbuf {

	var sourcePixbuf *gdk.Pixbuf

	sourcePixbuf = &data

	imagePixbuf := app.BackgroundImage.GetPixbuf()
	if imagePixbuf == nil {
		log.Println("Can't get on-screen pixbuf")
		return nil
	}

	allocation := widget.GetAllocation()
	// fmt.Printf("Width: %d/%d Height: %d/%d\n",
	//	allocation.GetWidth(),
	//	imagePixbuf.GetWidth(),
	//	allocation.GetHeight(),
	//	imagePixbuf.GetHeight())

	if allocation.GetWidth() != imagePixbuf.GetWidth() ||
		allocation.GetHeight() != imagePixbuf.GetHeight() {

		pixbuf, err := sourcePixbuf.ScaleSimple(
			allocation.GetWidth(),
			allocation.GetHeight(),
			gdk.INTERP_BILINEAR)
		if err != nil {
			log.Fatalln(err)
		}

		return pixbuf
	}

	return nil
}

func (app *App) getBackground(width, height int) *gdk.Pixbuf {

	if app.Config.Background.Source == "" {
		return nil
	}

	sourcePixbuf, err := gdk.PixbufNewFromFileAtScale(app.Config.Background.Source,
		width, height, app.Config.Background.PreserveAspectRatio)
	if err != nil {
		log.Fatal("Unable to load image:", err)
	}

	pixbuf, _ := gdk.PixbufCopy(sourcePixbuf)
	img, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		log.Fatalln(err)
	}
	app.BackgroundImage = img

	return sourcePixbuf
}

func (app App) getTerminal() *vteGtk.Terminal {

	terminal := vteGtk.NewTerminal()
	terminal.SetFontFromString(app.Config.Terminal.Font)

	terminal.SetEncoding("UTF-8")

	shell := os.Getenv("SHELL")

	// terminal.ExecSync("", []string{shell}, nil)
	terminal.ExecAsync(vte.Cmd{
		Args:    []string{shell},
		Timeout: -1,
		OnExec: func(pid int, err error) {
			if err != nil {
				log.Fatalln(err)
			}
		},
	})

	terminal.Connect("child-exited", func() {
		gtk.MainQuit()
	})

	terminal.Connect("window-title-changed", func() {
		app.Window.SetTitle(terminal.GetWindowTitle())
	})

	terminal.SetOpacity(app.Config.Terminal.Opacity)

	// color := gdk.NewRGBA()
	// color.SetRed(0.5)
	// color.SetGreen(0.0)
	// color.SetBlue(0.0)
	// terminal.SetBgColor(color)
	// terminal.SetFgColor(color)

	// fontDesc := pango.FontDescriptionNew()
	// fontDesc.SetSize(12)
	// fontDesc.SetFamily("Monospace")
	// terminal.SetFont(fontDesc)

	return terminal
}

func (app App) getTermWindow() *gtk.Window {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatalln(err)
	}
	win.Connect("destroy", gtk.MainQuit)
	win.SetTitle("TermLeek")
	win.SetDefaultSize(app.Config.Terminal.MinWidth, app.Config.Terminal.MinHeight)
	win.SetSizeRequest(app.Config.Terminal.MinWidth, app.Config.Terminal.MinHeight)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetResizable(true)
	if app.Config.Terminal.Icon != "" {
		win.SetIconFromFile(app.Config.Terminal.Icon)
	}

	return win
}

func (app *App) setupWindow() {

	app.Window = app.getTermWindow()

	// Get Image from file
	sourcePixbuf := app.getBackground(app.Config.Terminal.MinWidth, app.Config.Terminal.MinHeight)

	// Get New Terminal
	terminal := app.getTerminal()

	// Create New Scrolled Window
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	scrolledWindow.SetHExpand(true)
	scrolledWindow.SetVExpand(true)
	scrolledWindow.Add(terminal)

	// Create New Layout
	layout, err := gtk.LayoutNew(nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	app.Window.Add(layout)

	if sourcePixbuf != nil {
		layout.Put(app.BackgroundImage, 0, 0)

		// app.BackgroundImage.SetOpacity(0.5)
		app.BackgroundImage.SetSizeRequest(app.Config.Terminal.MinWidth, app.Config.Terminal.MinHeight)
	}

	scrolledWindow.SetSizeRequest(app.Config.Terminal.MinWidth, app.Config.Terminal.MinHeight)
	layout.Put(scrolledWindow, 0, 0)

	if sourcePixbuf != nil {
		app.Window.Connect("size-allocate", func(widget *gtk.Window) {
			if pixbuf := app.resizeImage(widget, *sourcePixbuf); pixbuf != nil {
				app.BackgroundImage.SetFromPixbuf(pixbuf)
				w := app.Window.GetAllocatedWidth()
				h := app.Window.GetAllocatedHeight()

				scrolledWindow.SetSizeRequest(w, h)
				terminal.SetSizeRequest(w, h)
			}
		})
	}
}

func NewApp() App {
	app := new(App)
	return *app
}

func main() {

	cfg, err := ini.Load("termleek.ini")
	if err != nil {
		fmt.Printf("Failed to read file: %v", err)
		os.Exit(1)
	}

	app := NewApp()

	app.Config.Background.Source = cfg.Section("Background").Key("source").MustString("")
	if app.Config.Background.Source != "" {
		if _, err := os.Stat(app.Config.Background.Source); errors.Is(err, os.ErrNotExist) {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	app.Config.Background.PreserveAspectRatio = cfg.Section("Background").Key("preserve_aspect_ratio").MustBool()

	app.Config.Terminal.Font = cfg.Section("Terminal").Key("font").MustString("monospace 10")

	app.Config.Terminal.MinWidth = func(w int) int {
		if w < minWidth {
			return minWidth
		}
		return w
	}(cfg.Section("Terminal").Key("min_width").MustInt(680))

	app.Config.Terminal.MinHeight = func(h int) int {
		if h < minHeight {
			return minHeight
		}
		return h
	}(cfg.Section("Terminal").Key("min_height").MustInt(370))

	// Forground Color
	// Background Color
	app.Config.Terminal.Opacity = cfg.Section("Terminal").Key("opacity").MustFloat64(1.0)
	app.Config.Terminal.Icon = cfg.Section("Terminal").Key("icon").MustString("")
	if app.Config.Terminal.Icon != "" {
		if _, err := os.Stat(app.Config.Terminal.Icon); errors.Is(err, os.ErrNotExist) {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	gtk.Init(nil)
	app.setupWindow()
	app.Window.ShowAll()
	gtk.Main()
}
