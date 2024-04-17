package show

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/common"
	"github.com/sayedmurtaza24/tinear/pkg/ui/views/dashboard"
)

type model struct {
	common *common.Model

	dashboard *dashboard.Model
}

func New(common *common.Model) *model {
	return &model{
		common:    common,
		dashboard: dashboard.New(common),
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

		m.common.Size.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		if key.Matches(msg, m.common.Keymap.Quit) {
			return m, tea.Quit
		}
	}

	d, cmd := m.dashboard.Update(msg)
	m.dashboard = d.(*dashboard.Model)

	return m, cmd
}

func (m *model) View() string {
	return m.dashboard.View()
}
