package issue

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

const (
	labelState    = "state:        "
	labelProject  = "project:      "
	labelTeam     = "team:         "
	labelAssignee = "assignee:     "
)

type Issue struct {
	Title    text.Focusable
	Desc     string
	Assignee text.Focusable
	State    text.Focusable
	Project  text.Focusable
	Team     text.Focusable
}

type Model struct {
	width   int
	height  int
	issue   Issue
	focused bool
}

type IssueOption func(*Model)

func WithFocused(focused bool) IssueOption {
	return func(m *Model) {
		m.focused = focused
	}
}

func New(w, h int, issue Issue, opts ...IssueOption) *Model {
	model := &Model{
		width:   w,
		height:  h,
		issue:   issue,
		focused: false,
	}

	for _, opt := range opts {
		opt(model)
	}

	return model
}

func (m *Model) Focused() bool {
	return m.focused
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	label := func(s string) string {
		t := text.Colored(s, color.Focusable("#777", "#444"))

		if m.focused {
			return t.Focused()
		} else {
			return t.Blurred()
		}
	}

	row := func(s text.Focusable) string {
		if m.focused {
			return s.Focused()
		} else {
			return s.Blurred()
		}
	}

	topBar := lipgloss.JoinVertical(
		lipgloss.Left,
		row(m.issue.Title),
		"",
		label(labelState)+row(m.issue.State),
		label(labelProject)+row(m.issue.Project),
		label(labelTeam)+row(m.issue.Team),
		label(labelAssignee)+row(m.issue.Assignee),
	)

	topBar = lipgloss.NewStyle().
		Padding(0, 1).Render(topBar)

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width),
	)

	description, _ := r.Render(m.issue.Desc)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		description,
	)
}
