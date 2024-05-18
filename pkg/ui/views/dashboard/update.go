package dashboard

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/linear/command"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case error:
		return m, tea.Quit

	case command.GetMeRes:
		m.state.Me = msg.Result

	case command.GetIssuesRes:
		m.table.SetLoading(false)

		m.state.Issues = append(m.state.Issues, msg.Result...)
		if msg.After != nil {
			cmds = append(cmds, command.GetIssues(m.client, m.store.ShouldReset(), msg.After))
		} else {
			if err := m.store.PutDiff(m.state.Issues...); err != nil {
				return m, tea.Quit
			}
		}

		m.renderTableRows(m.state.Issues)

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 5)
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
