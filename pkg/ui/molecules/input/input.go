package input

import (
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/ui/atoms/box"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Model struct {
	title       string
	options     []string
	input       textinput.Model
	suggestions table.Model
	width       int
}

func (m *Model) CurrentSuggestion() string {
	var current string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			recover()
		}()

		current = m.input.CurrentSuggestion()
	}()

	wg.Wait()

	return current
}

func makeSuggestionRows(opts []string) []*table.Row {
	rows := []*table.Row{}
	for _, opt := range opts {
		rows = append(rows, &table.Row{
			Identifier: opt,
			Items: []table.RowItem{
				{Normal: text.Plain(opt)},
			},
		})
	}

	return rows
}

func New(
	title string,
	prompt string,
	width, height int,
	options []string,
) *Model {
	var t table.Model

	input := textinput.New()

	if len(options) != 0 {
		input.SetSuggestions(options)
		input.ShowSuggestions = true

		rows := makeSuggestionRows(options)

		s := table.DefaultStyles()

		s.SelectedBlurred = lipgloss.NewStyle().Background(lipgloss.Color("2")).Bold(true)
		s.Header = lipgloss.NewStyle().Padding(0)

		t = table.New(
			table.WithStyles(s),
			table.WithColumns([]*table.Column{
				table.NewColumn(text.Plain("Options"), 1, table.WithAutoFill()),
			}),
			table.WithRows(rows),
			table.WithWidth(width-4),
			table.WithHeight(height-5),
			table.WithFocused(false),
			table.WithBackgroundColor("1", "#333"),
			table.WithNoHeader(),
		)
	}

	input.Width = width - 8

	input.Placeholder = prompt

	input.Prompt = " "

	input.KeyMap.NextSuggestion.SetHelp("ctrl+p", "Next")
	input.KeyMap.PrevSuggestion.SetHelp("ctrl+n", "Prev")
	input.KeyMap.AcceptSuggestion.SetHelp("tab", "Accept")

	input.PlaceholderStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#000")).
		Foreground(lipgloss.Color("#555"))

	return &Model{
		input:       input,
		title:       title,
		options:     options,
		width:       width,
		suggestions: t,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.input.Focus()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	var available []string
	for _, a := range m.options {
		if strings.HasPrefix(
			strings.ToLower(a),
			strings.ToLower(m.input.Value()),
		) {
			available = append(available, a)
		}
	}
	rows := makeSuggestionRows(available)
	m.suggestions.SetRows(rows)

	current := m.CurrentSuggestion()
	if current != "" {
		m.suggestions.SetSelectedRow(current)
	}

	if k, ok := msg.(tea.KeyMsg); ok && current != "" {
		if key.Matches(k, m.input.KeyMap.NextSuggestion) {
			m.suggestions.MoveDown(1)
		}
		if key.Matches(k, m.input.KeyMap.PrevSuggestion) {
			m.suggestions.MoveUp(1)
		}
	}

	m.suggestions, cmd = m.suggestions.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	input := lipgloss.NewStyle().
		Background(lipgloss.Color("#000")).
		Width(m.width - 4).
		Render(m.input.View())

	suggestions := m.suggestions.View()

	var content string
	if len(m.options) == 0 {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			input,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			input,
			"",
			suggestions,
		)
	}

	return box.New(
		"Assign",
		content,
		m.width,
		box.WithBorderStyle(
			lipgloss.NewStyle().
				Padding(1, 2, 1).
				BorderForeground(lipgloss.Color("#444")).
				Border(lipgloss.NormalBorder()),
		),
		box.WithLabelStyle(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("2")).
				Padding(0, 1).
				Bold(true),
		),
		box.WithBackground("#222"),
	)
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.input.KeyMap.NextSuggestion,
		m.input.KeyMap.PrevSuggestion,
		m.input.KeyMap.AcceptSuggestion,
	}
}

func (m *Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.input.KeyMap.NextSuggestion,
			m.input.KeyMap.PrevSuggestion,
			m.input.KeyMap.AcceptSuggestion,
		},
	}
}
