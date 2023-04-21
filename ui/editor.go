package ui

import (
	"github.com/jroimartin/gocui"
)

var noActionEditor gocui.Editor = gocui.EditorFunc(func(*gocui.View, gocui.Key, rune, gocui.Modifier) {})
