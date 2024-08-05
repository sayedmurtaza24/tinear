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

const (
	FocusIssues focus = iota
	FocusProjects
	FocusVisual
	FocusSort
	FocusFilter
	FocusHover
)

const (
	ViewAll view = iota
	ViewProject
)

var focusNextMap = map[focus][]focus{
	FocusProjects: {FocusIssues},
	FocusIssues:   {FocusVisual, FocusSort, FocusFilter, FocusHover},
}

type (
	focus          int
	focusStackItem struct {
		mode  focus
		onPop tea.Cmd
	}
	focusStack []focusStackItem

	view  int
	Model struct {
		width   int
		height  int
		syncing bool

		focus focusStack

		currView  view
		switching bool

		hovered *store.Issue

		store  *store.Store
		client *client.Client

		prjTable table.Model
		table    table.Model
		input    textinput.Model

		err error
	}
)

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

	model.table = table.New(
		table.WithFocused(true),
		table.WithSpinner(spinner.Dot),
		table.WithLoadingText("loading..."),
		table.WithVisualMode(true),
		table.WithStyles(st),
	)
	model.prjTable = table.New(
		table.WithFocused(false),
		table.WithStyles(st),
	)

	model.client = client
	model.store = store
	model.syncing = true

	model.input = textinput.New()
	model.input.Prompt = ""
	model.input.TextStyle = model.input.TextStyle.Foreground(lipgloss.Color("#999"))

	model.focus = []focusStackItem{{mode: FocusIssues}}

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

func (s *focusStack) current() focus {
	return (*s)[len(*s)-1].mode
}

func (s *focusStack) push(m focus, onPop tea.Cmd) bool {
	focusableNext, ok := focusNextMap[(*s)[len(*s)-1].mode]
	if !ok {
		return false
	}

	for _, allowed := range focusableNext {
		if allowed != m {
			continue
		}
		*s = append(*s, focusStackItem{
			mode:  m,
			onPop: onPop,
		})
		return true
	}
	return false
}

func (s *focusStack) pop() tea.Cmd {
	if len(*s) == 1 {
		return nil
	}
	cleanup := (*s)[len(*s)-1].onPop
	*s = (*s)[:len(*s)-1]
	return cleanup
}
