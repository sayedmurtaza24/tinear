package box

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type styleType int

const (
	styleTypeBoxStyle styleType = iota
	styleTypeLabelStyle
	styleTypeBg
)

type boxOption func() (lipgloss.Style, styleType)

func WithLabelStyle(s lipgloss.Style) boxOption {
	return func() (lipgloss.Style, styleType) {
		return s, styleTypeLabelStyle
	}
}

func WithBorderStyle(s lipgloss.Style) boxOption {
	return func() (lipgloss.Style, styleType) {
		return s, styleTypeBoxStyle
	}
}

func WithBackground(bgColor string) boxOption {
	return func() (lipgloss.Style, styleType) {
		return lipgloss.NewStyle().Background(lipgloss.Color(bgColor)), styleTypeBg
	}
}

func New(
	label, content string,
	width int,
	styles ...boxOption,
) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1)

	labelStyle := lipgloss.NewStyle().
		PaddingTop(0).
		PaddingBottom(0).
		PaddingLeft(1).
		PaddingRight(1)

	bgStyle := lipgloss.NewStyle()

	for _, s := range styles {
		style, bType := s()
		switch bType {
		case styleTypeBoxStyle:
			boxStyle = style
		case styleTypeLabelStyle:
			labelStyle = style
		case styleTypeBg:
			bgStyle = style
		}
	}

	var (
		border = boxStyle.
			BorderBackground(bgStyle.GetBackground()).
			GetBorderStyle()

		topBorderStyler = lipgloss.NewStyle().
				Foreground(boxStyle.GetBorderTopForeground()).
				Background(bgStyle.GetBackground()).
				Render

		topLeft  = topBorderStyler(border.TopLeft)
		topRight = topBorderStyler(border.TopRight)

		renderedLabel = labelStyle.Inherit(bgStyle).Render(label)
	)

	borderWidth := boxStyle.GetHorizontalBorderSize()
	cellsShort := max(0, width+borderWidth-lipgloss.Width(topLeft+topRight+renderedLabel))
	gap := strings.Repeat(border.Top, cellsShort)
	top := lipgloss.JoinHorizontal(
		lipgloss.Left,
		topLeft,
		renderedLabel,
		topBorderStyler(gap),
		topRight,
	)

	bottom := boxStyle.Copy().
		Inherit(bgStyle).
		BorderTop(false).
		Width(width).
		Render(content)

	return bgStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, top, bottom),
	)
}
