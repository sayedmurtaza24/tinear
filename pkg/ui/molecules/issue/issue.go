package issue

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Model struct {
	title  string
	desc   string
	width  int
	height int
}

func New(
	title,
	desc,
	assignee string,
	state,
	projectName,
	teamName text.Focusable,
	w, h int,
) *Model {
	return &Model{
		width:  w,
		height: h,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	return ""
}
