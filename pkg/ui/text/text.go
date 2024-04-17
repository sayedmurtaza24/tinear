package text

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
)

const (
	leftArrow  = ""
	rightArrow = ""
)

type Focusable interface {
	Raw() string
	Focused() string
	Blurred() string
}

type Opt int

const (
	// Bold
	B Opt = iota
	// Italic
	I
)

type plain struct {
	value    string
	rendered string
}

func Plain(value string, opts ...Opt) plain {
	render := lipgloss.NewStyle()
	for _, opt := range opts {
		if opt == B {
			render = render.Bold(true)
		}
		if opt == I {
			render = render.Italic(true)
		}
	}

	return plain{
		value:    value,
		rendered: render.Render(value),
	}
}

func (p plain) Raw() string     { return p.value }
func (p plain) Focused() string { return p.value }
func (p plain) Blurred() string { return p.value }

type colored struct {
	raw     string
	focused string
	blurred string
}

func Colored(value string, fg color.Color, opts ...Opt) colored {
	focused := lipgloss.NewStyle().Foreground(fg.Focused())
	blurred := lipgloss.NewStyle().Foreground(fg.Blurred())

	for _, opt := range opts {
		if opt == B {
			focused = focused.Bold(true)
			blurred = blurred.Bold(true)
		}
		if opt == I {
			focused = focused.Italic(true)
			blurred = blurred.Italic(true)
		}
	}

	str := colored{
		raw:     value,
		focused: focused.Render(value),
		blurred: blurred.Render(value),
	}
	return str
}

func (s colored) Raw() string     { return s.raw }
func (s colored) Focused() string { return s.focused }
func (s colored) Blurred() string { return s.blurred }

type chip struct {
	raw     string
	focused string
	blurred string
}

func Chip(value string, fg, bg color.Color, opts ...Opt) colored {
	focused := lipgloss.NewStyle().
		Background(bg.Focused()).
		Foreground(fg.Focused()).
		Padding(0, 1)
	blurred := lipgloss.NewStyle().
		Background(bg.Blurred()).
		Foreground(fg.Blurred()).
		Padding(0, 1)

	for _, opt := range opts {
		if opt == B {
			focused = focused.Bold(true)
			blurred = blurred.Bold(true)
		}
		if opt == I {
			focused = focused.Italic(true)
			blurred = blurred.Italic(true)
		}
	}

	return colored{
		raw:     value,
		focused: focused.Render(value),
		blurred: blurred.Render(value),
	}
}

func (s chip) Raw() string     { return s.raw }
func (s chip) Focused() string { return s.focused }
func (s chip) Blurred() string { return s.blurred }
