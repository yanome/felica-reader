package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/yanome/felica-reader/felica"
	"github.com/yanome/felica-reader/usb"
)

const (
	LEFT_WIDTH    = 10
	CENTER_WIDTH  = 30
	RIGHT_WIDTH   = 55
	TOP_HEIGHT    = 10
	MIDDLE_HEIGHT = 10
	BOTTOM_HEIGHT = 5
)

type ui struct {
	menu        *menu
	log         *log
	systems     *list
	services    *list
	decoded     *text
	raw         *text
	initialized bool
}

func Init(g *gocui.Gui, e *usb.Endpoints) error {
	var card *felica.Card
	system := 0
	service := 0
	u := &ui{}
	u.menu = &menu{
		view: &view{
			name:  "menu",
			title: "(M)enu",
			key:   'm',
			sizeFn: func(_, maxY int) viewSize {
				return viewSize{0, 0, LEFT_WIDTH, maxY - BOTTOM_HEIGHT - 2}
			},
		},
		options: []menuOption{
			{
				text: "Read",
				callback: func() error {
					card = &felica.Card{}
					if err := u.loadCard(g, e, card); err != nil {
						card = nil
						return u.log.Print(g, fmt.Sprintf("Read error: %s", err))
					}
					system = 0
					service = 0
					return nil
				},
			},
			{
				text: "Save",
				callback: func() error {
					if err := u.saveCard(g, card); err != nil {
						return u.log.Print(g, fmt.Sprintf("Save error: %s", err))
					}
					return nil
				},
			},
			{
				text: "Quit",
				callback: func() error {
					return gocui.ErrQuit
				},
			},
		},
	}
	u.systems = &list{
		view: &view{
			name:  "systems",
			title: "S(y)stems",
			key:   'y',
			sizeFn: func(_, _ int) viewSize {
				return viewSize{LEFT_WIDTH + 1, 0, LEFT_WIDTH + CENTER_WIDTH + 1, TOP_HEIGHT}
			},
		},
		callback: func(idx int) error {
			if idx == system {
				return nil
			}
			system = idx
			service = 0
			return u.selectSystem(g, card, idx)
		},
	}
	u.services = &list{
		view: &view{
			name:  "services",
			title: "S(e)rvices",
			key:   'e',
			sizeFn: func(_, maxY int) viewSize {
				return viewSize{LEFT_WIDTH + 1, TOP_HEIGHT + 1, LEFT_WIDTH + CENTER_WIDTH + 1, maxY - BOTTOM_HEIGHT - 2}
			},
		},
		callback: func(idx int) error {
			if idx == service {
				return nil
			}
			service = idx
			return u.selectService(g, card, system, idx)
		},
	}
	u.decoded = &text{
		view: &view{
			name:  "decoded",
			title: "(D)ecoded contents",
			key:   'd',
			sizeFn: func(maxX, _ int) viewSize {
				return viewSize{LEFT_WIDTH + CENTER_WIDTH + 2, 0, maxX - 1, TOP_HEIGHT}
			},
		},
	}
	u.raw = &text{
		view: &view{
			name:  "raw",
			title: "(R)aw contents",
			key:   'r',
			sizeFn: func(maxX, maxY int) viewSize {
				return viewSize{LEFT_WIDTH + CENTER_WIDTH + 2, TOP_HEIGHT + 1, maxX - 1, maxY - BOTTOM_HEIGHT - 2}
			},
		},
	}
	u.log = &log{
		view: &view{
			name:  "log",
			title: "Log",
			key:   0,
			sizeFn: func(maxX, maxY int) viewSize {
				return viewSize{0, maxY - BOTTOM_HEIGHT - 1, maxX - 1, maxY - 1}
			},
		},
	}
	g.SetManager(u)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		return fmt.Errorf("failed to set key binding: %s", err)
	}
	return nil
}

func (u *ui) Layout(g *gocui.Gui) error {
	if maxX, maxY := g.Size(); maxX < LEFT_WIDTH+CENTER_WIDTH+RIGHT_WIDTH+3 || maxY < TOP_HEIGHT+MIDDLE_HEIGHT+BOTTOM_HEIGHT+3 {
		return fmt.Errorf("terminal is too small")
	}
	for _, section := range []section{u.menu, u.systems, u.services, u.decoded, u.raw, u.log} {
		if err := section.update(g); err != nil {
			return fmt.Errorf("error updating section: %s", err)
		}
	}
	if !u.initialized {
		u.initialized = true
		if err := u.menu.view.setActive(g); err != nil {
			return err
		}
	}
	return nil
}

func (u *ui) loadCard(g *gocui.Gui, e *usb.Endpoints, c *felica.Card) error {
	if err := u.systems.Clean(g); err != nil {
		return err
	}
	if err := u.services.Clean(g); err != nil {
		return err
	}
	if err := u.decoded.Clean(g); err != nil {
		return err
	}
	if err := u.raw.Clean(g); err != nil {
		return err
	}
	card, err := felica.Read(e)
	if err != nil {
		return fmt.Errorf("error reading card: %s", err)
	}
	if card == nil {
		return fmt.Errorf("no card found")
	}
	*c = *card
	u.log.Print(g, fmt.Sprintf("Found card %s", c))
	systems := []string{}
	for _, system := range c.Systems {
		systems = append(systems, system.String())
	}
	if err = u.systems.UpdateList(g, systems, 0); err != nil {
		return err
	}
	if len(systems) > 0 {
		if err = u.selectSystem(g, c, 0); err != nil {
			return err
		}
	}
	return nil
}

func (u *ui) saveCard(g *gocui.Gui, c *felica.Card) error {
	if c == nil {
		return fmt.Errorf("no card loaded")
	}
	jsonBytes, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %s", err)
	}
	fileName := fmt.Sprintf("%s.json", c)
	if err = os.WriteFile(fileName, jsonBytes, 0o600); err != nil {
		return fmt.Errorf("error writing file: %s", err)
	}
	u.log.Print(g, fmt.Sprintf("Card information saved to file %s", fileName))
	return nil
}

func (u *ui) selectSystem(g *gocui.Gui, c *felica.Card, system int) error {
	services := []string{}
	for _, service := range c.Systems[system].Services {
		services = append(services, service.String())
	}
	if err := u.services.UpdateList(g, services, 0); err != nil {
		return err
	}
	if len(services) > 0 {
		if err := u.selectService(g, c, system, 0); err != nil {
			return err
		}
	} else {
		if err := u.decoded.Clean(g); err != nil {
			return err
		}
		if err := u.raw.Clean(g); err != nil {
			return err
		}
	}
	return nil
}

func (u *ui) selectService(g *gocui.Gui, c *felica.Card, system int, service int) error {
	if err := u.decoded.Clean(g); err != nil {
		return err
	}
	if err := u.raw.Clean(g); err != nil {
		return err
	}
	if err := u.decoded.Printf(g, "%s", c.Systems[system].DecodedService(service)); err != nil {
		return err
	}
	if err := u.raw.Printf(g, "%s", c.Systems[system].RawService(service)); err != nil {
		return err
	}
	return nil
}
