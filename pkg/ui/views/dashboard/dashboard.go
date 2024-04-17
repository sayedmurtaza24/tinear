package dashboard

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/common"
	"github.com/sayedmurtaza24/tinear/pkg/ui/atoms/help"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/layouts"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/input"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/status"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Model struct {
	common         *common.Model
	status         *status.Model
	loadingStatus  *status.Status
	input          *input.Model
	loadingSpinner spinner.Model
	table          table.Model
}

func New(common *common.Model) *Model {
	// withColor := func(s string) string {
	// 	return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(s)
	// }

	st := status.New(common)

	s := table.DefaultStyles()

	s.Selected = lipgloss.NewStyle().
		Background(lipgloss.Color("#2e2e2e")).
		Bold(true).
		Border(lipgloss.BlockBorder(), false).
		BorderLeftForeground(lipgloss.Color("#22aaee")).
		BorderLeft(true)

	s.SelectedBlurred = lipgloss.NewStyle().
		Background(lipgloss.Color("#2e2e2e")).
		Bold(true).
		Border(lipgloss.BlockBorder(), false).
		BorderLeftForeground(lipgloss.Color("24")).
		BorderLeft(true)

	s.Header = s.Header.
		Padding(1, 1, 0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	t := table.New(
		table.WithColumns([]*table.Column{
			table.NewColumn(text.Plain("Column 1"), 0.5),
			table.NewColumn(text.Plain("Column 2"), 0.5, table.WithAutoFill()),
			table.NewColumn(text.Plain("Column 3"), 0.1),
			table.NewColumn(text.Plain("Column 4"), 0.1),
			table.NewColumn(text.Plain("Column 5"), 0.1),
		}),
		table.WithFocused(false),
		table.WithStyles(s),
		table.WithSpinner(spinner.MiniDot),
		table.WithLoadingText("Loading issues..."),
		table.WithHeight(28),
		table.WithWidth(50),
		table.WithVisualMode(true),
	)

	var rows []*table.Row

	for i := 1; i < 133; i++ {
		t := "Normal Row " + fmt.Sprint(i)
		rows = append(rows, &table.Row{
			Identifier: fmt.Sprintf("%d", i),
			Items: []table.RowItem{
				{
					Normal:   text.Colored(t, color.Focusable("#ff22ee", "#aaa").Darken(0.3), text.B),
					Selected: text.Colored(t, color.Focusable("#ffaa2b", "#aaaa2b")),
				},
				{
					Normal:   text.Colored(t, color.Focusable("#ff22ee", "#aaa")),
					Selected: text.Colored(t, color.Focusable("#ffaa2b", "#aaaa2b")),
				},
				{
					Normal:   text.Colored(t, color.Focusable("#ddd", "#aaa")),
					Selected: text.Colored(t, color.Focusable("#ffaa2b", "#aaaa2b")),
				},
				{
					Normal:   text.Colored(t, color.Focusable("#ddd", "#aaa")),
					Selected: text.Colored(t, color.Focusable("#ffaa2b", "#aaaa2b")),
				},
				{
					Normal:   text.Colored(t, color.Focusable("#ddd", "#aaa")),
					Selected: text.Colored(t, color.Focusable("#ffaa2b", "#aaaa2b")),
				},
			},
		})
	}

	t.SetRows(rows)

	loadingSp := spinner.New(
		spinner.WithSpinner(spinner.Moon),
	)

	st1 := status.NewStatus(
		text.Plain(loadingSp.View()+" Loading"),
		status.Right,
	)
	st2 := status.NewStatus(text.Plain("Arbitrary Text"), status.Right)
	st3 := status.NewStatus(text.Plain("A chip here"), status.Center)
	st4 := status.NewStatus(text.Plain("Another chip here"), status.Center)

	st.SetStatuses([]*status.Status{st1, st2, st3, st4})

	input := input.New(
		"Assign Task",
		"Enter assignee name",
		30,
		17,
		[]string{
			"Dollar",
			"Bitcoin",
			"Ethereum",
			"Euro",
			"Ruble",
			"Yen",
			"Pound",
			"Yuan",
			"Krone",
			"Yuan",
			"Toman",
			"Peso",
			"Franc",
		},
	)

	return &Model{
		common:         common,
		table:          t,
		status:         st,
		loadingStatus:  st1,
		loadingSpinner: loadingSp,
		input:          input,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.status.Init(),
		m.loadingSpinner.Tick,
		m.input.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "right" {
			m.table.SetWidth(m.table.Width() + 1)
		}
		if key.String() == "left" {
			m.table.SetWidth(m.table.Width() - 1)
		}
	}

	_, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	m.loadingStatus.SetStatus(text.Plain(m.loadingSpinner.View() + " Loading..."))

	bg := m.table.View()

	list := layouts.PlaceOverlay(
		layouts.Center,
		m.input.View(),
		bg,
	)

	h := help.New(m.table.KeyMap, m.common.Size.Width())

	return lipgloss.JoinVertical(lipgloss.Left, list, m.status.View(), h)
}

func (m *Model) ShortHelp() []key.Binding {
	return m.table.KeyMap.ShortHelp()
}

func (m *Model) FullHelp() [][]key.Binding {
	return m.table.KeyMap.FullHelp()
}
