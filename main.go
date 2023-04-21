package main

import (
	"log"

	"github.com/jroimartin/gocui"
	"github.com/yanome/felica-reader/ui"
	"github.com/yanome/felica-reader/usb"
)

func main() {
	endpoints, release, err := usb.Init()
	if err != nil {
		log.Panicf("Failed to initialize USB: %s", err)
	}
	defer release()
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicf("Failed to create UI: %s", err)
	}
	defer g.Close()
	if err = ui.Init(g, endpoints); err != nil {
		log.Panicf("Failed to initialize UI: %s", err)
	}
	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicf("Unexpected error: %s", err)
	}
}
