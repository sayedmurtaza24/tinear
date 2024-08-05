package show

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/views/dashboard"
)

type model struct {
	dashboard *dashboard.Model
}

func New(store *store.Store, client *client.Client) *model {
	return &model{
		dashboard: dashboard.New(store, client),
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
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
