package status

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/common"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Mode string

const (
	Normal Mode = "NORMAL"
	Visual Mode = "VISUAL"
)

type Alignment int

const (
	Left Alignment = iota
	Right
	Center
)

type Status struct {
	status text.Focusable

	alignment Alignment
}

func (s *Status) SetStatus(st text.Focusable) {
	s.status = st
}

func NewStatus(st text.Focusable, align Alignment) *Status {
	s := &Status{
		status:    st,
		alignment: align,
	}

	return s
}

func (m *Model) renderStatuses(w int, focused bool) string {
	getText := func(t text.Focusable) string {
		if !focused {
			return t.Blurred()
		}

		return t.Focused()
	}

	var out []string

	var lefts []*Status
	var centers []*Status
	var rights []*Status

	totalWidth := 0
	for _, st := range m.statuses {
		totalWidth += lipgloss.Width(getText(st.status))
		switch st.alignment {
		case Left:
			lefts = append(lefts, st)
		case Center:
			centers = append(centers, st)
		case Right:
			rights = append(rights, st)
		}
	}

	left := w - totalWidth

	if left < 0 {
		return lipgloss.NewStyle().Inline(true).MaxWidth(w).Render(
			lipgloss.JoinHorizontal(lipgloss.Left, out...),
		)
	}

	spaceBetween := left / 2
	leftover := left % 2

	pad := func(w int) string {
		return lipgloss.NewStyle().Width(w).Background(m.styles.Bar.GetBackground()).Render(" ")
	}

	for _, l := range lefts {
		out = append(out, lipgloss.NewStyle().
			Background(m.styles.Bar.GetBackground()).
			Inline(true).
			Render(getText(l.status)),
		)
	}

	out = append(out, pad(spaceBetween))

	for _, c := range centers {
		out = append(out, lipgloss.NewStyle().
			Background(m.styles.Bar.GetBackground()).
			Inline(true).
			Render(getText(c.status)),
		)
	}

	out = append(out, pad(spaceBetween+leftover))

	for _, r := range rights {
		out = append(out, lipgloss.NewStyle().
			Background(m.styles.Bar.GetBackground()).
			Inline(true).
			Render(getText(r.status)),
		)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, out...)
}

type Styles struct {
	ModeIndicator map[Mode]lipgloss.Style
	Bar           lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		ModeIndicator: map[Mode]lipgloss.Style{
			Normal: lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				Background(lipgloss.Color("210")),
			Visual: lipgloss.NewStyle().Bold(true).Padding(0, 1),
		},
		Bar: lipgloss.NewStyle().
			Background(lipgloss.Color("240")),
	}
}

type Model struct {
	common *common.Model

	styles Styles

	mode Mode

	focused bool

	statuses []*Status
}

type modelOpt func(*Model)

func WithStyles(s Styles) modelOpt {
	return func(m *Model) {
		m.styles = s
	}
}

func WithMode(mode Mode) modelOpt {
	return func(m *Model) {
		m.mode = mode
	}
}

func New(common *common.Model, opts ...modelOpt) *Model {
	m := &Model{
		common:  common,
		mode:    Normal,
		styles:  DefaultStyles(),
		focused: true,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

type StatusTicker time.Time

func (m *Model) Tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return time.Now()
	})
}

func (m *Model) SetStatuses(statuses []*Status) {
	m.statuses = statuses
}

func (m *Model) SetMode(mode Mode) {
	m.mode = mode
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) Init() tea.Cmd {
	return m.Tick()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case StatusTicker:
		return m, m.Tick()
	}
	return m, nil
}

func (m *Model) View() string {
	// modeStyle := m.styles.ModeIndicator[m.mode]

	// mode := chip.New(
	// 	string(m.mode),
	// 	chip.WithBold(),
	// 	chip.WithBgColor(modeStyle.GetBackground()),
	// 	chip.WithFgColor(modeStyle.GetForeground()),
	// 	chip.WithPadding(modeStyle.GetHorizontalPadding()/2),
	// 	chip.WithRightArrow(m.styles.Bar.GetBackground()),
	// )

	// modeW := lipgloss.Width(mode)

	// st := m.renderStatuses(m.common.Size.Width()-modeW, m.focused)

	// bar := m.styles.Bar.MaxWidth(m.common.Size.Width()).Render(
	// 	lipgloss.JoinHorizontal(lipgloss.Left, mode, st),
	// )

	// return bar
	return ""
}

func (m *Model) ShortHelp() []key.Binding {
	return nil
}

func (m *Model) FullHelp() [][]key.Binding {
	return nil
}
