package common

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/keymap"
	"github.com/sayedmurtaza24/tinear/pkg/screen"
)

type Model struct {
	Size   *screen.Size
	Keymap *keymap.KeyMap

	tea.Model
	help.KeyMap
}

func New(keymap *keymap.KeyMap, size *screen.Size) *Model {
	return &Model{
		Size:   size,
		Keymap: keymap,
	}
}
