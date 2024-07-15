package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

type bgColor struct {
	Normal  string
	Blurred string
}

type selectedRange struct {
	start int
	end   int
}

func (s *selectedRange) SetStart(i int) {
	s.start = i
}

func (s *selectedRange) SetEnd(i int) {
	s.end = i
}

func (s *selectedRange) InRange(i int) bool {
	if s.start == s.end {
		return i == s.start
	}

	if s.end > s.start {
		return i >= s.start && i <= s.end
	}

	return i >= s.end && i <= s.start
}

func (s *selectedRange) GetRange() (start, end int) {
	if s.start > s.end {
		return s.end, s.start
	}
	return s.start, s.end
}

type Model struct {
	styles Styles

	KeyMap KeyMap

	backgroundColor bgColor
	loadingText     string

	cols []*Column
	rows []*Row

	spinner      spinner.Spinner
	spinnerModel spinner.Model

	itemsHeight int
	itemsWidth  int

	cursor int
	start  int

	focus          bool
	loading        bool
	shouldTickNext bool

	visualModeEnabled bool
	visualMode        bool
	selectedRange     selectedRange

	noHeader bool
}

type RowItem struct {
	Normal   text.Focusable
	Selected text.Focusable
}

type Row struct {
	Identifier string
	Items      []RowItem
}

type Column struct {
	title       text.Focusable
	maxWidth    int
	minWidth    int
	widthFactor float32
	fill        bool

	calculatedWidth int
}

type columnOption func(*Column)

func WithMaxWidth(width int) columnOption {
	return func(c *Column) {
		c.maxWidth = width
	}
}

func WithMinWidth(width int) columnOption {
	return func(c *Column) {
		c.minWidth = width
	}
}

func WithAutoFill() columnOption {
	return func(c *Column) {
		c.fill = true
	}
}

func NewColumn(title text.Focusable, widthFactor float32, opts ...columnOption) *Column {
	col := Column{
		title: title,
	}

	col.widthFactor = min(max(widthFactor, 0.0), 1.0)

	for _, opt := range opts {
		opt(&col)
	}

	return &col
}

type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	VisualMode   key.Binding
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.LineUp,
		km.LineDown,
		km.VisualMode,
		km.GotoTop,
		km.HalfPageUp,
		km.HalfPageDown,
	}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.LineUp, km.LineDown, km.GotoTop, km.GotoBottom},
		{km.HalfPageUp, km.HalfPageDown, km.VisualMode},
	}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "end"),
		),
		VisualMode: key.NewBinding(
			key.WithKeys("v", "V"),
			key.WithHelp("shift+v/v", "visual"),
		),
	}
}

type Styles struct {
	Header lipgloss.Style

	Selected lipgloss.Style

	SelectedBlurred lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Selected:        lipgloss.NewStyle().Background(lipgloss.Color("#333")).Padding(0, 1).Bold(true),
		SelectedBlurred: lipgloss.NewStyle().Background(lipgloss.Color("#222")).Padding(0, 1),
		Header:          lipgloss.NewStyle().Bold(true).Padding(0, 1),
	}
}

func (m *Model) SetStyles(s Styles) {
	m.styles = s
}

type Option func(*Model)

func New(opts ...Option) Model {
	m := Model{
		cursor: 0,

		spinnerModel: spinner.New(),
		spinner:      spinner.Dot,
		KeyMap:       DefaultKeyMap(),
		styles:       DefaultStyles(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func WithNoHeader() Option {
	return func(m *Model) {
		m.noHeader = true
	}
}

func WithColumns(cols []*Column) Option {
	return func(m *Model) {
		m.cols = cols
		m.calculateColsWidth()
	}
}

func WithRows(rows []*Row) Option {
	return func(m *Model) {
		m.rows = rows
	}
}

func WithHeight(h int) Option {
	return func(m *Model) {
		m.itemsHeight = h - lipgloss.Height(m.headersView())
	}
}

func WithWidth(w int) Option {
	return func(m *Model) {
		m.itemsWidth = w - 2
		m.calculateColsWidth()
	}
}

func WithFocused(f bool) Option {
	return func(m *Model) {
		m.focus = f
	}
}

func WithStyles(s Styles) Option {
	return func(m *Model) {
		m.styles = s
	}
}

func WithKeyMap(km KeyMap) Option {
	return func(m *Model) {
		m.KeyMap = km
	}
}

func WithSpinner(sp spinner.Spinner) Option {
	return func(m *Model) {
		m.spinner = sp
		m.spinnerModel.Spinner = sp
	}
}

func WithIsLoading(b bool) Option {
	return func(m *Model) {
		m.loading = b
		m.shouldTickNext = b
	}
}

func WithLoadingText(text string) Option {
	return func(m *Model) {
		m.loadingText = text
	}
}

func WithBackgroundColor(normal string, blurred string) Option {
	return func(m *Model) {
		m.backgroundColor = bgColor{Normal: normal, Blurred: blurred}
	}
}

func WithVisualMode(b bool) Option {
	return func(m *Model) {
		m.visualModeEnabled = b
	}
}

func scale(values []*Column) {
	total := sum(values)
	scale := 1.0 / total
	for i := range values {
		values[i].widthFactor = values[i].widthFactor * scale
	}
}

func sum(values []*Column) float32 {
	var sum float32
	for _, v := range values {
		sum += v.widthFactor
	}
	return sum
}

func (m *Model) calculateColsWidth() {
	scale(m.cols)

	_, r, _, l := m.styles.Header.GetPadding()
	colPadding := r + l

	for _, col := range m.cols {
		v := int(float32(m.itemsWidth) * col.widthFactor)

		if !col.fill && col.maxWidth > 0 {
			v = min(v, col.maxWidth)
		}

		if !col.fill && col.minWidth > 0 {
			v = max(v, col.minWidth)
		}

		col.calculatedWidth = max(v-colPadding, 0)
	}

	leftover := func() int {
		used := 0
		for _, col := range m.cols {
			if col.calculatedWidth < 1 {
				continue
			}
			used += col.calculatedWidth + colPadding
		}
		return m.itemsWidth - used
	}

	adjustableCols := func(leftover int) []*Column {
		adjustable := []*Column{}
		for _, col := range m.cols {
			if col.fill && leftover > 0 {
				adjustable = append(adjustable, col)
			}
			if col.fill && leftover < 0 && col.calculatedWidth > 0 {
				adjustable = append(adjustable, col)
			}
		}
		return adjustable
	}

	left := leftover()
	cols := adjustableCols(left)
	for left != 0 && len(cols) > 0 {
		left = leftover()
		for _, col := range cols {
			if col.fill && left > 0 {
				col.calculatedWidth++
				left--
			}
			if col.fill && left < 0 {
				col.calculatedWidth--
				left++
			}
		}
		cols = adjustableCols(left)
	}
}

func (m *Model) DebugView() string {
	var out []string

	out = append(out, "Table Debug View")
	out = append(out, "Width: "+fmt.Sprint(m.itemsWidth+2))

	for _, col := range m.cols {
		out = append(out, fmt.Sprintf("%s = %d", col.title.Raw(), col.calculatedWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.shouldTickNext {
		m.shouldTickNext = false
		return m, m.spinnerModel.Tick
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinnerModel, cmd = m.spinnerModel.Update(msg)
		return m, cmd
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.itemsHeight / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.itemsHeight / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, m.KeyMap.VisualMode):
			m.SetVisualMode(!m.visualMode)
		case msg.Type == tea.KeyEsc:
			m.visualMode = false
		}
	}

	return m, nil
}

func (m Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) TopOffset() int {
	return m.cursor - m.start + lipgloss.Height(m.headersView())
}

func (m *Model) Loading() bool {
	return m.loading
}

func (m *Model) SetLoading(b bool) tea.Cmd {
	m.loading = b

	return m.spinnerModel.Tick
}

func (m Model) View() string {
	header := m.headersView()

	var headerH int
	if !m.noHeader {
		headerH = lipgloss.Height(header)
	}

	content := make([]string, 0, headerH+m.itemsHeight)

	if !m.noHeader {
		content = append(content, header)
	}

	if m.loading {
		sp := lipgloss.JoinHorizontal(lipgloss.Left, m.spinnerModel.View(), " ", m.loadingText)
		spinner := lipgloss.NewStyle().
			Width(m.itemsWidth).
			Align(lipgloss.Center).
			MarginTop(m.itemsHeight / 2).Render(sp)

		content = append(content, spinner)
	} else {
		end := clamp(m.start+m.itemsHeight, 0, len(m.rows))
		for i := m.start; i < end; i++ {
			content = append(content, m.renderRow(i))
		}
	}

	if m.focus {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(m.backgroundColor.Normal)).
			Height(m.itemsHeight + headerH).
			Width(m.itemsWidth + 2).
			Render(lipgloss.JoinVertical(lipgloss.Center, content...))
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(m.backgroundColor.Blurred)).
		Height(m.itemsHeight + headerH).
		Width(m.itemsWidth + 2).
		Render(lipgloss.JoinVertical(lipgloss.Center, content...))
}

func (m Model) SelectedRow() (identifier string) {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}

	return m.rows[m.cursor].Identifier
}

func (m Model) VisualMode() bool {
	return m.visualMode
}

func (m Model) SelectedRows() (identifiers []string) {
	if !m.visualMode {
		return []string{m.SelectedRow()}
	}

	start, end := m.selectedRange.GetRange()

	var rows []string

	for i := start; i <= end; i++ {
		rows = append(rows, m.rows[i].Identifier)
	}

	return rows
}

func (m Model) Rows() []*Row {
	return m.rows
}

func (m Model) Columns() []*Column {
	return m.cols
}

func (m *Model) SetSelectedRow(identifier string) {
	for i, r := range m.rows {
		if r.Identifier == identifier {
			m.SetCursor(i)
			return
		}
	}
}

func (m *Model) SetRows(r []*Row) {
	m.rows = r

	m.SetCursor(m.cursor)
}

func (m *Model) SetColumns(c []*Column) {
	m.cols = c
	m.calculateColsWidth()
}

func (m *Model) SetWidth(w int) {
	m.itemsWidth = w - 2
	m.calculateColsWidth()
}

func (m *Model) SetHeight(h int) {
	m.itemsHeight = h - lipgloss.Height(m.headersView())
}

func (m Model) Height() int {
	return m.itemsHeight + lipgloss.Height(m.headersView())
}

func (m Model) Width() int {
	return m.itemsWidth + 2
}

func (m Model) Cursor() int {
	return m.cursor
}

func (m *Model) SetCursor(n int) {
	m.cursor = clamp(n, 0, len(m.rows)-1)
}

func (m *Model) SetVisualMode(b bool) {
	if m.visualModeEnabled {
		m.visualMode = b
		m.selectedRange.SetStart(m.cursor)
		m.selectedRange.SetEnd(m.cursor)
	}
}

func (m *Model) MoveUp(n int) {
	m.cursor = clamp(m.cursor-n, 0, len(m.rows)-1)
	if m.visualModeEnabled {
		if !m.visualMode {
			m.selectedRange.SetStart(m.cursor)
		}
		m.selectedRange.SetEnd(m.cursor)
	}

	if m.cursor < m.start {
		m.start = clamp(m.start-n, 0, len(m.rows)-m.itemsHeight)
	}
}

func (m *Model) MoveDown(n int) {
	m.cursor = clamp(m.cursor+n, 0, len(m.rows)-1)
	if m.visualModeEnabled {
		if !m.visualMode {
			m.selectedRange.SetStart(m.cursor)
		}
		m.selectedRange.SetEnd(m.cursor)
	}

	if m.cursor >= m.itemsHeight+m.start {
		m.start = clamp(m.start+n, 0, len(m.rows)-m.itemsHeight)
	}
}

func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

func (m *Model) GotoBottom() {
	m.MoveDown(len(m.rows))
}

func (m Model) headersView() string {
	if m.noHeader {
		return ""
	}

	var color lipgloss.Color
	if m.focus {
		color = lipgloss.Color(m.backgroundColor.Normal)
	} else {
		color = lipgloss.Color(m.backgroundColor.Blurred)
	}

	s := make([]string, 0, len(m.cols))

	for _, col := range m.cols {
		if col.calculatedWidth <= 0 {
			continue
		}

		value := col.title.Focused()
		if !m.focus {
			value = col.title.Blurred()
		}

		var v string

		w := lipgloss.Width(value)

		if w > col.calculatedWidth {
			v = truncate.StringWithTail(value, uint(col.calculatedWidth), "…")
		} else {
			v = lipgloss.JoinHorizontal(
				lipgloss.Left,
				value,
				lipgloss.NewStyle().
					Background(color).
					Render(strings.Repeat(" ", col.calculatedWidth-w)),
			)
		}

		v = m.styles.Header.
			Background(color).
			BorderBackground(color).
			Render(v)

		s = append(s, v)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, s...)
}

func (m *Model) fill(w int, selected bool) string {
	if selected {
		if m.focus {
			return lipgloss.NewStyle().
				Background(lipgloss.Color(m.backgroundColor.Normal)).
				Background(m.styles.Selected.GetBackground()).
				Width(w).
				Render(" ")
		} else {
			return lipgloss.NewStyle().
				Background(lipgloss.Color(m.backgroundColor.Blurred)).
				Background(m.styles.SelectedBlurred.GetBackground()).
				Width(w).
				Render(" ")
		}
	}

	if m.focus {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(m.backgroundColor.Normal)).
			Width(w).
			Render(" ")
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color(m.backgroundColor.Blurred)).
		Width(w).
		Render(" ")
}

func (m *Model) style(str string, selected bool) string {
	if selected {
		if m.focus {
			return lipgloss.NewStyle().
				SetString(str).
				Background(lipgloss.Color(m.backgroundColor.Normal)).
				Background(m.styles.Selected.GetBackground()).
				Bold(m.styles.Selected.GetBold()).
				Italic(m.styles.Selected.GetItalic()).
				String()
		} else {
			return lipgloss.NewStyle().
				SetString(str).
				Background(lipgloss.Color(m.backgroundColor.Blurred)).
				Background(m.styles.SelectedBlurred.GetBackground()).
				Bold(m.styles.SelectedBlurred.GetBold()).
				Italic(m.styles.SelectedBlurred.GetItalic()).
				String()
		}
	}

	if m.focus {
		return lipgloss.NewStyle().
			SetString(str).
			Background(lipgloss.Color(m.backgroundColor.Normal)).
			String()
	}

	return lipgloss.NewStyle().
		SetString(str).
		Background(lipgloss.Color(m.backgroundColor.Blurred)).
		String()
}

func (m *Model) border(str string, selected bool) string {
	if selected {
		if m.focus {
			return lipgloss.NewStyle().
				SetString(str).
				Border(m.styles.Selected.GetBorder()).
				BorderLeftBackground(m.styles.Selected.GetBorderLeftBackground()).
				BorderRightBackground(m.styles.Selected.GetBorderRightBackground()).
				BorderTopBackground(m.styles.Selected.GetBorderTopBackground()).
				BorderBottomBackground(m.styles.Selected.GetBorderBottomBackground()).
				BorderLeftForeground(m.styles.Selected.GetBorderLeftForeground()).
				BorderRightForeground(m.styles.Selected.GetBorderRightForeground()).
				BorderTopForeground(m.styles.Selected.GetBorderTopForeground()).
				BorderBottomForeground(m.styles.Selected.GetBorderBottomForeground()).
				Padding(m.styles.Selected.GetPadding()).
				String()
		} else {
			return lipgloss.NewStyle().
				SetString(str).
				Border(m.styles.SelectedBlurred.GetBorder()).
				BorderLeftBackground(m.styles.SelectedBlurred.GetBorderLeftBackground()).
				BorderRightBackground(m.styles.SelectedBlurred.GetBorderRightBackground()).
				BorderTopBackground(m.styles.SelectedBlurred.GetBorderTopBackground()).
				BorderBottomBackground(m.styles.SelectedBlurred.GetBorderBottomBackground()).
				BorderLeftForeground(m.styles.SelectedBlurred.GetBorderLeftForeground()).
				BorderRightForeground(m.styles.SelectedBlurred.GetBorderRightForeground()).
				BorderTopForeground(m.styles.SelectedBlurred.GetBorderTopForeground()).
				BorderBottomForeground(m.styles.SelectedBlurred.GetBorderBottomForeground()).
				Padding(m.styles.SelectedBlurred.GetPadding()).
				String()
		}
	}

	return lipgloss.NewStyle().Render(str)
}

func (m *Model) renderRow(rowID int) string {
	r := m.styles.Header.GetPaddingRight()
	l := m.styles.Header.GetPaddingLeft()

	s := make([]string, 0, len(m.cols))
	for i, v := range m.rows[rowID].Items {
		if m.cols[i].calculatedWidth <= 0 {
			continue
		}

		curr := m.cursor == rowID || (m.visualMode && m.selectedRange.InRange(rowID))

		var text text.Focusable
		if curr && v.Selected != nil && v.Selected.Raw() != "" {
			text = v.Selected
		} else {
			text = v.Normal
		}

		var value string
		if m.focus {
			value = text.Focused()
		} else {
			value = text.Blurred()
		}

		value = m.style(value, curr)

		w := lipgloss.Width(value)
		if w >= m.cols[i].calculatedWidth {
			value = lipgloss.JoinHorizontal(
				lipgloss.Left,
				m.fill(l, curr),
				truncate.StringWithTail(value, uint(m.cols[i].calculatedWidth), "…"),
				m.fill(r, curr),
			)
		} else {
			value = lipgloss.JoinHorizontal(
				lipgloss.Left,
				m.fill(l, curr),
				value,
				m.fill(m.cols[i].calculatedWidth-w, curr),
				m.fill(r, curr),
			)
		}

		s = append(s, value)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left, s...)

	row = m.border(row, m.cursor == rowID)

	return row
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
