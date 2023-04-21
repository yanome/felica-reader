package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

type section interface {
	update(g *gocui.Gui) error
}

type menu struct {
	view    *view
	options []menuOption
}

type menuOption struct {
	text     string
	callback func() error
}

func (s *menu) update(g *gocui.Gui) error {
	if s.view.exists(g) {
		if err := s.view.update(g); err != nil {
			return err
		}
	} else {
		view, err := s.view.create(g)
		if err != nil {
			return err
		}
		last := len(s.options) - 1
		for idx, option := range s.options {
			fmt.Fprint(view, option.text)
			if idx != last {
				fmt.Fprintln(view)
			}
		}
		if err = g.SetKeybinding(s.view.name, gocui.KeyEnter, gocui.ModNone, func(_ *gocui.Gui, v *gocui.View) error {
			_, cy := v.Cursor()
			_, oy := v.Origin()
			return s.options[cy+oy].callback()
		}); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
		if err = g.SetKeybinding(s.view.name, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
		if err = g.SetKeybinding(s.view.name, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
	}
	return nil
}

type list struct {
	view     *view
	callback func(idx int) error
}

func (s *list) update(g *gocui.Gui) error {
	if s.view.exists(g) {
		if err := s.view.update(g); err != nil {
			return err
		}
	} else {
		if _, err := s.view.create(g); err != nil {
			return err
		}
		if err := g.SetKeybinding(s.view.name, gocui.KeyEnter, gocui.ModNone, func(_ *gocui.Gui, v *gocui.View) error {
			lines := v.BufferLines()
			if len(lines) > 0 {
				_, cy := v.Cursor()
				_, oy := v.Origin()
				for idx := range lines {
					if slices := strings.SplitN(lines[idx], "", 2); len(slices) == 2 {
						lines[idx] = slices[1]
					} else {
						lines[idx] = ""
					}
				}
				if err := s.UpdateList(g, lines, cy+oy); err != nil {
					return err
				}
				if err := v.SetCursor(0, cy); err != nil {
					return fmt.Errorf("error updating cursor: %s", err)
				}
				if err := v.SetOrigin(0, oy); err != nil {
					return fmt.Errorf("error updating origin: %s", err)
				}
				return s.callback(cy + oy)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
		if err := g.SetKeybinding(s.view.name, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
		if err := g.SetKeybinding(s.view.name, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
	}
	return nil
}

func (s *list) UpdateList(g *gocui.Gui, options []string, active int) error {
	view, err := s.view.get(g)
	if err != nil {
		return err
	}
	view.Clear()
	if err = view.SetCursor(0, 0); err != nil {
		return fmt.Errorf("error updating cursor: %s", err)
	}
	if err = view.SetOrigin(0, 0); err != nil {
		return fmt.Errorf("error updating origin: %s", err)
	}
	last := len(options) - 1
	for idx, option := range options {
		marker := " "
		if idx == active {
			marker = "»"
		}
		fmt.Fprintf(view, "%s%s", marker, option)
		if idx != last {
			fmt.Fprintln(view)
		}
	}
	return nil
}

func (s *list) Clean(g *gocui.Gui) error {
	return s.UpdateList(g, []string{}, -1)
}

type text struct {
	view *view
}

func (s *text) update(g *gocui.Gui) error {
	if s.view.exists(g) {
		if err := s.view.update(g); err != nil {
			return err
		}
	} else {
		if _, err := s.view.create(g); err != nil {
			return err
		}
		if err := g.SetKeybinding(s.view.name, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
		if err := g.SetKeybinding(s.view.name, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
			return fmt.Errorf("failed to set key binding: %s", err)
		}
	}
	return nil
}

func (s *text) Printf(g *gocui.Gui, format string, a ...any) error {
	view, err := s.view.get(g)
	if err != nil {
		return err
	}
	fmt.Fprintf(view, format, a...)
	return nil
}

func (s *text) Clean(g *gocui.Gui) error {
	view, err := s.view.get(g)
	if err != nil {
		return err
	}
	view.Clear()
	return nil
}

type log struct {
	view *view
}

func (s *log) update(g *gocui.Gui) error {
	if s.view.exists(g) {
		if err := s.view.update(g); err != nil {
			return err
		}
	} else {
		view, err := s.view.create(g)
		if err != nil {
			return err
		}
		view.Autoscroll = true
	}
	return nil
}

func (s *log) Print(g *gocui.Gui, message string) error {
	view, err := s.view.get(g)
	if err != nil {
		return err
	}
	fmt.Fprintf(view, "\n%s %s", time.Now().Format("15:04:05"), message)
	return nil
}
