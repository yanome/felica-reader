package ui

import (
	"fmt"
	"unicode"

	"github.com/jroimartin/gocui"
)

type view struct {
	name   string
	title  string
	key    rune
	sizeFn func(maxX, maxY int) viewSize
}

type viewSize struct {
	x0, y0, x1, y1 int
}

func (v *view) get(g *gocui.Gui) (*gocui.View, error) {
	view, err := g.View(v.name)
	if err != nil {
		return nil, fmt.Errorf("cannot find view %s: %s", v.name, err)
	}
	return view, nil
}

func (v *view) exists(g *gocui.Gui) bool {
	_, err := v.get(g)
	return err == nil
}

func (v *view) update(g *gocui.Gui) error {
	size := v.sizeFn(g.Size())
	view, err := g.SetView(v.name, size.x0, size.y0, size.x1, size.y1)
	if err != nil {
		return fmt.Errorf("error udpating view %s: %s", v.name, err)
	}
	_, maxY := view.Size()
	_, cy := view.Cursor()
	_, oy := view.Origin()
	if cy >= maxY {
		y := 1 + cy - maxY
		if err = view.SetCursor(0, cy-y); err != nil {
			return fmt.Errorf("error updating cursor: %s", err)
		}
		if err = view.SetOrigin(0, oy+y); err != nil {
			return fmt.Errorf("error updating origin: %s", err)
		}
	}
	return nil
}

func (v *view) create(g *gocui.Gui) (*gocui.View, error) {
	size := v.sizeFn(g.Size())
	view, err := g.SetView(v.name, size.x0, size.y0, size.x1, size.y1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, fmt.Errorf("error creating view %s: %s", v.name, err)
		}
		view.Title = v.title
		view.SelFgColor = gocui.AttrReverse
		view.Editor = noActionEditor
		if err = v.setKey(g); err != nil {
			return nil, err
		}
	}
	return view, nil
}

func (v *view) setKey(g *gocui.Gui) error {
	if v.key != 0 {
		for _, k := range []rune{
			unicode.ToLower(v.key),
			unicode.ToUpper(v.key),
		} {
			if err := g.SetKeybinding("", k, gocui.ModNone, func(g *gocui.Gui, _ *gocui.View) error {
				return v.setActive(g)
			}); err != nil {
				return fmt.Errorf("failed to set key binding: %s", err)
			}
		}
	}
	return nil
}

func (v *view) setActive(g *gocui.Gui) error {
	if view := g.CurrentView(); view != nil {
		view.Highlight = false
	}
	view, err := g.SetCurrentView(v.name)
	if err != nil {
		return fmt.Errorf("error setting active view: %s", err)
	}
	view.Highlight = true
	return nil
}

func cursorDown(_ *gocui.Gui, v *gocui.View) error {
	return moveCursor(v, 1)
}

func cursorUp(_ *gocui.Gui, v *gocui.View) error {
	return moveCursor(v, -1)
}

func moveCursor(v *gocui.View, y int) error {
	_, maxY := v.Size()
	_, cy := v.Cursor()
	_, oy := v.Origin()
	cy += y
	if cy < 0 {
		if oy > 0 {
			if err := v.SetOrigin(0, oy-1); err != nil {
				return fmt.Errorf("error updating origin: %s", err)
			}
		}
	} else if oy+cy < len(v.BufferLines()) {
		if cy >= maxY {
			if err := v.SetOrigin(0, oy+1); err != nil {
				return fmt.Errorf("error updating origin: %s", err)
			}
		} else {
			if err := v.SetCursor(0, cy); err != nil {
				return fmt.Errorf("error updating cursor: %s", err)
			}
		}
	}
	return nil
}
