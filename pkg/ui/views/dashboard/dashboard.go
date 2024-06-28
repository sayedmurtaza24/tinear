package dashboard

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
	"github.com/sayedmurtaza24/tinear/pkg/common"
	"github.com/sayedmurtaza24/tinear/pkg/linear/command"
	"github.com/sayedmurtaza24/tinear/pkg/linear/sort"
	"github.com/sayedmurtaza24/tinear/pkg/storage"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/input"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/issue"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/status"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
)

type Model struct {
	state DashboardState

	store storage.IssueStore

	client linearClient.LinearClient

	sortOption sort.SortOption

	loadingStatus *status.Status

	common  *common.Model
	status  *status.Model
	input   *input.Model
	spinner spinner.Model
	table   table.Model
	issue   *issue.Model
}

func New(common *common.Model, store storage.IssueStore, client linearClient.LinearClient) *Model {
	var state DashboardState
	var model Model

	st := table.DefaultStyles()

	st.Selected = st.Selected.
		Border(lipgloss.NormalBorder(), false).
		BorderForeground(lipgloss.Color("#2D4F67")).
		UnsetPadding().
		BorderLeft(true)

	st.Header = st.Header.
		Border(lipgloss.ThickBorder(), false).
		BorderForeground(lipgloss.Color("#3d3223")).
		BorderBottom(true)

	t := table.New(
		table.WithFocused(true),
		table.WithSpinner(spinner.Dot),
		table.WithLoadingText("loading..."),
		table.WithVisualMode(true),
		table.WithStyles(st),
		table.WithIsLoading(len(store.Get()) == 0),
	)
	model.table = t

	model.client = client
	model.common = common
	model.state = state
	model.store = store

	return &model
}

func (m *Model) Init() tea.Cmd {
	m.renderTableCols(false)
	m.renderTableRows(m.store.Get())

	return tea.Batch(
		command.GetMe(m.client),
		command.GetIssues(m.client, m.store, nil),
	)
}

func (m *Model) ShortHelp() []key.Binding {
	return m.table.KeyMap.ShortHelp()
}

func (m *Model) FullHelp() [][]key.Binding {
	return m.table.KeyMap.FullHelp()
}
