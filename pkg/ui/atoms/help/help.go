package help

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

func New(keys help.KeyMap, width int, showAll bool) string {
	h := help.New()

	h.ShowAll = showAll

	h.Width = width

	horizontalPadding := 2
	verticalPadding := 1

	if !showAll {
		verticalPadding = 0
	}

	return lipgloss.NewStyle().Padding(verticalPadding, horizontalPadding).Render(h.View(keys))
}
