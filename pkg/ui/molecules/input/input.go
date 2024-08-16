package input

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type Suggestion struct {
	Identifier string
	// WARN: has to be unique
	Title string
	Color string
}

type Model struct {
	title            string
	options          []Suggestion
	input            textinput.Model
	suggestions      table.Model
	width, maxHeight int
}

func makeSuggestionRows(opts []Suggestion) []*table.Row {
	rows := []*table.Row{}

	for _, opt := range opts {
		var rowNormal text.Focusable
		if opt.Color != "" {
			rowNormal = text.Colored(opt.Title, color.Simple(opt.Color))
		} else {
			rowNormal = text.Plain(opt.Title)
		}

		rows = append(rows, &table.Row{
			Identifier: opt.Identifier,
			Items:      []table.RowItem{{Normal: rowNormal}},
		})
	}

	return rows
}

func New(
	prompt string,
	width, maxHeight int,
	options bool,
) Model {
	var t table.Model

	input := textinput.New()

	if options {
		s := table.DefaultStyles()

		s.SelectedBlurred = lipgloss.NewStyle().Background(lipgloss.Color("#333")).Bold(true)
		s.Header = lipgloss.NewStyle().Padding(0)

		t = table.New(
			table.WithStyles(s),
			table.WithColumns([]*table.Column{
				table.NewColumn(text.Plain("Options"), 1, table.WithAutoFill()),
			}),
			table.WithWidth(width),
			table.WithFocused(false),
			table.WithNoHeader(),
		)
	}

	input.Prompt = " "
	input.Placeholder = prompt
	input.PlaceholderStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#000")).
		Foreground(lipgloss.Color("#555"))
	input.Focus()

	return Model{
		input:       input,
		width:       width,
		maxHeight:   maxHeight,
		suggestions: t,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Reset() {
	m.input.SetValue("")
	m.suggestions.SetCursor(0)
}

func (m *Model) Highlighted() *Suggestion {
	for _, sug := range m.options {
		if sug.Identifier == m.suggestions.SelectedRow() {
			return &sug
		}
	}
	return nil
}

func (m *Model) Value() string {
	return strings.TrimSpace(m.input.Value())
}

func (m *Model) SetSuggestions(suggestions []Suggestion) {
	m.options = suggestions
	m.suggestions.SetRows(makeSuggestionRows(suggestions))
	m.suggestions.SetHeight(min(len(suggestions)+1, m.maxHeight))
}

func (m *Model) SetPlaceholder(placeholder string) {
	m.input.Placeholder = placeholder
}

func (m *Model) SetValue(value string) {
	m.input.SetValue(value)
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.input.Width = width - 2
	m.suggestions.SetWidth(width)
}

func (m *Model) filterSuggestions() {
	var available []Suggestion
	for _, a := range m.options {
		if strings.HasPrefix(
			strings.ToLower(a.Title),
			strings.ToLower(m.input.Value()),
		) {
			available = append(available, a)
		}
	}
	rows := makeSuggestionRows(available)
	m.suggestions.SetRows(rows)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.filterSuggestions()

	current := m.Highlighted()
	if current != nil {
		m.input.ShowSuggestions = true
		m.input.SetSuggestions([]string{current.Title})
	} else {
		m.input.ShowSuggestions = false
		m.suggestions.SetCursor(0)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+n" {
			m.suggestions.MoveDown(1)
		}
		if msg.String() == "ctrl+p" {
			m.suggestions.MoveUp(1)
		}
	}

	m.suggestions, cmd = m.suggestions.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	input := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#000")).
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
			suggestions,
		)
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("#333")).
		Render(content)
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
