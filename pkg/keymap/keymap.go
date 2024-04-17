package keymap

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit key.Binding
}

func NewDefault() *KeyMap {
	return &KeyMap{
		Quit: key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("ctrl+c, q", "quit")),
	}
}
