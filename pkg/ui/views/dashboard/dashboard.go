package dashboard

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/input"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/issue"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/status"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
)

type Model struct {
	width  int
	height int

	store *store.Store

	client *client.Client

	loadingStatus *status.Status

	status  *status.Model
	input   *input.Model
	spinner spinner.Model
	table   table.Model
	issue   *issue.Model
}

func New(store *store.Store, client *client.Client) *Model {
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
		// table.WithIsLoading(len(store.Get()) == 0),
	)
	model.table = t

	model.client = client
	model.store = store

	return &model
}

func (m *Model) Init() tea.Cmd {
	m.renderTableCols(false)
	// m.renderTableRows(m.store.Get())

	return tea.Batch(
		m.client.GetMe(),
		m.client.GetIssues(nil),
	)
}
