package dashboard

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/linear/resumable"
	"github.com/sayedmurtaza24/tinear/pkg/linear/user"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case error:
		return m, tea.Quit

	case GetMeResponse:
		m.state.Me = user.User(msg.User)
		m.state.OrganizationName = msg.OrganizationName

	case resumable.Command[GetMyIssuesResponse]:
		log.Println("GetMyIssuesResponse")
		m.table.SetLoading(false)

		m.state.MyIssues = append(m.state.MyIssues, msg.Result...)
		if msg.After != nil {
			cmds = append(cmds, GetMyIssues(m.client, m.issuesSortOption, msg.After))
		}

		m.renderTableRows(m.state.MyIssues)

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 5)
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
