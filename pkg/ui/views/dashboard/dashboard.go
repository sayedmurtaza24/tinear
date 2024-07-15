package dashboard

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
)

type Model struct {
	width      int
	height     int
	syncing    bool
	sortMode   bool
	filterMode bool

	hovered *store.Issue

	store  *store.Store
	client *client.Client

	table table.Model
	input textinput.Model
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
	)
	model.table = t

	model.client = client
	model.store = store
	model.syncing = true

	model.input = textinput.New()
	model.input.Prompt = ""

	return &model
}

func (m *Model) Init() tea.Cmd {
	m.updateTableCols()

	return tea.Batch(
		m.updateIssues(),
		m.table.SetLoading(m.store.Current().FirstTime),
		m.client.GetOrg(),
	)
}
