package text

import (
	"strings"

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

type Opt func(style lipgloss.Style) lipgloss.Style

var B Opt = func(style lipgloss.Style) lipgloss.Style {
	return style.Bold(true)
}

var I Opt = func(style lipgloss.Style) lipgloss.Style {
	return style.Italic(true)
}

var Width = func(width int) Opt {
	return func(style lipgloss.Style) lipgloss.Style {
		return style.Width(width)
	}
}

type plain struct {
	value    string
	rendered string
}

func Plain(value string, opts ...Opt) plain {
	render := lipgloss.NewStyle()
	for _, opt := range opts {
		opt(render)
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
		opt(focused)
		opt(blurred)
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
		opt(focused)
		opt(blurred)
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

type keymapText struct {
	raw     string
	focused string
	blurred string
}

func KeymapText(
	value string,
	fgColor color.Color,
	keymapIndex int,
	keymapColor color.Color,
	opts ...Opt,
) keymapText {
	focused := lipgloss.NewStyle().Foreground(fgColor.Focused())
	blurred := lipgloss.NewStyle().Foreground(fgColor.Blurred())

	hfocused := lipgloss.NewStyle().Foreground(keymapColor.Focused())
	hblurred := lipgloss.NewStyle().Foreground(keymapColor.Blurred())

	for _, opt := range opts {
		opt(focused)
		opt(blurred)
		opt(hfocused)
		opt(hblurred)
	}

	hfocusedStr := strings.Join([]string{
		focused.Render(value[:keymapIndex]),
		hfocused.Render(string(value[keymapIndex])),
		focused.Render(value[keymapIndex+1:]),
	}, "")

	str := keymapText{
		raw:     value,
		blurred: blurred.Render(value),
		focused: hfocusedStr,
	}

	return str
}

func (k keymapText) Raw() string {
	return k.raw
}

func (k keymapText) Focused() string {
	return k.focused
}

func (k keymapText) Blurred() string {
	return k.blurred
}
