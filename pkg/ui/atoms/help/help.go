package help

import (
	"github.com/charmbracelet/bubbles/help"
)

func New(keys help.KeyMap, width int) string {
	h := help.New()

	h.ShowAll = false

	h.Width = width

	return h.View(keys)
}
