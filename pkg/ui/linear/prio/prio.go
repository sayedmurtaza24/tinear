package prio

import (
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Prio int

func (p Prio) ToFocusable(brighten float64) text.Focusable {
	return prioColors(p, brighten)
}

func prioColors(p Prio, brighten float64) text.Focusable {
	switch p {
	case 1:
		return text.Colored("Urgent", color.Focusable("#e03a43", "#888").Brighten(brighten), text.B)
	case 2:
		return text.Colored("High", color.Focusable("#d47248", "#888").Brighten(brighten))
	case 3:
		return text.Colored("Medium", color.Focusable("#806b38", "#888").Brighten(brighten))
	case 4:
		return text.Colored("Low", color.Focusable("#4a4a4a", "#888").Brighten(brighten))
	}

	return text.Plain("")
}
