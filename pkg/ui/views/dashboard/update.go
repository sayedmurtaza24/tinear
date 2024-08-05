package dashboard

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

const projectsTableWidth = 25

func forceUpdate() tea.Msg {
	return struct{}{}
}

func returnError(err error) tea.Cmd {
	return func() tea.Msg { return err }
}

func (m *Model) handleSortMode(key tea.KeyMsg) tea.Cmd {
	switch m.focus.current() {
	case FocusIssues:
		if key.String() != "s" {
			return nil
		}
		onPop := tea.Batch(
			func() tea.Msg {
				m.table.Focus()
				m.updateTableCols()
				return nil
			},
			m.updateIssues(),
		)

		if m.focus.push(FocusSort, onPop) {
			m.table.Blur()
			m.updateTableCols()
		}

	case FocusSort:
		mappings := map[string]store.SortMode{
			"p": store.SortModeProject,
			"t": store.SortModeTitle,
			"a": store.SortModeAssignee,
			"e": store.SortModeState,
			"r": store.SortModePrio,
			"g": store.SortModeAge,
			"m": store.SortModeTeam,
			"s": store.SortModeSmart,
		}

		sortMode, ok := mappings[key.String()]
		if !ok {
			return nil
		}

		err := m.store.SetSortMode(sortMode)
		if err != nil {
			return returnError(err)
		}

		return m.focus.pop()
	}
	return nil
}

func (m *Model) handleFilter(key tea.KeyMsg) (cmd tea.Cmd) {
	switch m.focus.current() {
	case FocusIssues:
		onPop := func() tea.Msg {
			if m.input.Value() == "/" {
				m.input.SetValue("")
			}
			m.input.Blur()
			return m.updateIssues()()
		}

		if key.String() == "/" {
			if m.focus.push(FocusFilter, onPop) {
				m.input.SetValue("")
				m.input.Focus()
				m.input, cmd = m.input.Update(key)
				return cmd
			}
		}
	case FocusFilter:
		if key.Type == tea.KeyEnter {
			return m.focus.pop()
		}

		m.input, cmd = m.input.Update(key)

		return tea.Batch(cmd, m.updateIssues(withDebounce()))
	}
	return nil
}

func (m *Model) handleBookmark(key tea.KeyMsg) tea.Cmd {
	if m.focus.current() != FocusIssues {
		return nil
	}

	if key.String() != "b" {
		return nil
	}

	selectedIssues := m.table.SelectedRows()

	err := m.store.ToggleBookmark(selectedIssues...)
	if err != nil {
		return returnError(err)
	}

	m.table.SetVisualMode(false)

	var selected string
	if len(selectedIssues) == 1 {
		selected = selectedIssues[0]
	}

	return m.updateIssues(withSelected(selected))
}

func (m *Model) handleOpen(key tea.KeyMsg) tea.Cmd {
	if m.focus.current() != FocusIssues && m.focus.current() != FocusHover {
		return nil
	}

	if key.String() != "o" {
		return nil
	}

	issue, err := m.store.Issue(m.table.SelectedRow())
	if err != nil {
		return returnError(err)
	}

	open := func(url string) error {
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "windows":
			cmd = "cmd"
			args = []string{"/c", "start"}
		case "darwin":
			cmd = "open"
		default:
			cmd = "xdg-open"
		}
		args = append(args, url)
		return exec.Command(cmd, args...).Start()
	}

	linearBaseURL := "https://linear.app"
	urlKey := m.store.Current().Org.URLKey

	open(fmt.Sprintf("%s/%s/issue/%s", linearBaseURL, urlKey, issue.Identifier))

	return nil
}

func (m *Model) handleHover(key tea.KeyMsg) tea.Cmd {
	switch m.focus.current() {
	case FocusIssues:
		if key.String() != "K" {
			return nil
		}
		onPop := func() tea.Msg {
			m.hovered = nil
			m.table.Focus()
			return forceUpdate()
		}
		if m.focus.push(FocusHover, onPop) {
			issue, err := m.store.Issue(m.table.SelectedRow())
			if err != nil {
				return returnError(err)
			}
			m.hovered = issue
			m.table.Blur()
		}

	case FocusHover:
		if key.String() == "o" || key.Type == tea.KeyEsc {
			return nil
		}
		return m.focus.pop()
	}

	return nil
}

func (m *Model) handleClose(key tea.KeyMsg) tea.Cmd {
	if m.focus.current() != FocusIssues {
		return nil
	}

	if key.String() != "q" {
		return nil
	}

	return tea.Quit
}

func (m *Model) handleFocusStack(key tea.KeyMsg) tea.Cmd {
	if key.Type != tea.KeyEscape {
		return nil
	}
	if m.input.Value() != "" {
		m.input.SetValue("")
		return m.updateIssues()
	}
	return m.focus.pop()
}

func (m *Model) handleProjectSelection(key tea.KeyMsg) tea.Cmd {
	if m.focus.current() != FocusProjects {
		return nil
	}

	if key.Type != tea.KeyEnter {
		return nil
	}

	projects, err := m.store.Projects()
	if err != nil {
		return returnError(err)
	}

	var selectedProject, currProject *store.Project
	for _, prj := range projects {
		if prj.ID == m.prjTable.SelectedRow() {
			selectedProject = &prj
		}
	}
	if selectedProject == nil {
		// NOTE: shouldn't really happen
		return tea.Quit
	}

	currProject = m.store.Current().Project

	if currProject == nil || currProject.ID != selectedProject.ID {
		m.store.SetProject(selectedProject)
	}

	shiftFocus := func() tea.Msg {
		m.prjTable.Focus()
		m.table.Blur()
		return nil
	}
	onPop := tea.Batch(shiftFocus, m.updateIssues())

	if m.focus.push(FocusIssues, onPop) {
		m.prjTable.Blur()
		m.table.Focus()
	}

	return m.updateIssues()
}

func (m *Model) handleViews(key tea.KeyMsg) tea.Cmd {
	if key.Type != tea.KeyTab {
		return nil
	}

	if m.focus.current() != FocusIssues && m.focus.current() != FocusProjects {
		return nil
	}

	m.table.SetLoading(true)

	projects, err := m.store.Projects()
	if err != nil {
		return returnError(err)
	}

	switch m.currView {
	case ViewAll:
		m.currView = ViewProject
		m.focus = []focusStackItem{{mode: FocusProjects}}

		if len(projects) > 0 {
			m.store.SetProject(&projects[0])
		}
		m.table.SetWidth(m.width - projectsTableWidth)
		m.prjTable.SetWidth(projectsTableWidth)
		m.prjTable.Focus()
		m.table.Blur()
		m.prjTable.SetOnMove(func(selectedID string) tea.Cmd {
			var selectedProject *store.Project
			for _, prj := range projects {
				if prj.ID == selectedID {
					selectedProject = &prj
				}
			}
			if selectedProject != nil {
				m.store.SetProject(selectedProject)
			}
			return m.updateIssues(withCursorAt(0), withColumnsUpdate())
		})

		return m.updateIssues(withColumnsUpdate())
	case ViewProject:
		m.currView = ViewAll
		m.focus = []focusStackItem{{mode: FocusIssues}}

		m.store.SetProject(nil)
		m.table.SetWidth(m.width)
		m.table.Focus()
		m.prjTable.Blur()
		m.prjTable.SetOnMove(nil)

		return m.updateIssues(withColumnsUpdate())
	}

	return nil
}

type (
	updateIssuesMsg struct {
		issues        []store.Issue
		selected      string
		cursor        int
		updateColumns func()
	}
	issueUpdateOpt struct {
		selected      string
		cursor        int
		debounce      bool
		updateColumns bool
	}
	issueUpdateOptFunc func(opt *issueUpdateOpt)
)

func withSelected(selected string) issueUpdateOptFunc {
	return func(opt *issueUpdateOpt) {
		opt.selected = selected
	}
}

func withDebounce() issueUpdateOptFunc {
	return func(opt *issueUpdateOpt) {
		opt.debounce = true
	}
}

func withCursorAt(cursor int) issueUpdateOptFunc {
	return func(opt *issueUpdateOpt) {
		opt.cursor = cursor
	}
}

func withColumnsUpdate() issueUpdateOptFunc {
	return func(opt *issueUpdateOpt) {
		opt.updateColumns = true
	}
}

func (m *Model) updateIssues(opts ...issueUpdateOptFunc) tea.Cmd {
	options := issueUpdateOpt{cursor: -1}

	for _, opt := range opts {
		opt(&options)
	}

	return func() tea.Msg {
		var issues []store.Issue
		var err error

		currentInput := m.input.Value()

		if options.debounce {
			time.Sleep(150 * time.Millisecond)

			if currentInput != m.input.Value() {
				return nil
			}
		}

		if len(m.input.Value()) > 3 {
			issues, err = m.store.SearchIssues(m.input.Value()[1:])
			if err != nil {
				if errors.Is(err, store.ErrNoOrgSelected) {
					return nil
				}
				return err
			}
		} else {
			issues, err = m.store.Issues()
			if err != nil {
				if errors.Is(err, store.ErrNoOrgSelected) {
					return nil
				}
				return err
			}
		}

		var projects []store.Project

		if options.updateColumns {
			projects, err = m.store.Projects()
			if err != nil {
				return err
			}
		}

		updateCols := func() {
			if options.updateColumns {
				m.updateTableCols()
				m.updateProjectsTable(projects)
			}
		}

		return updateIssuesMsg{
			issues:        issues,
			selected:      options.selected,
			cursor:        options.cursor,
			updateColumns: updateCols,
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case error:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		cmds = append(cmds, m.handleFilter(msg))
		cmds = append(cmds, m.handleBookmark(msg))
		cmds = append(cmds, m.handleHover(msg))
		cmds = append(cmds, m.handleOpen(msg))
		cmds = append(cmds, m.handleSortMode(msg))
		cmds = append(cmds, m.handleClose(msg))
		cmds = append(cmds, m.handleFocusStack(msg))
		cmds = append(cmds, m.handleProjectSelection(msg))
		cmds = append(cmds, m.handleViews(msg))

	case updateIssuesMsg:
		msg.updateColumns()
		m.updateTableRows(msg.issues)
		if msg.selected != "" {
			m.table.SetSelectedRow(msg.selected)
		} else if msg.cursor != -1 {
			m.table.SetCursor(msg.cursor)
		}
		m.table.SetLoading(false)

	case client.GetOrgRes:
		changed, err := m.store.StoreOrg(msg.Result)
		if err != nil {
			return m, returnError(err)
		}

		if changed {
			m.syncing = true
			m.table.SetLoading(true)

			cmds = append(cmds, tea.Batch(
				m.client.GetProjects(nil),
				m.client.GetTeams(nil),
				m.client.GetUsers(nil),
				m.client.GetIssues(m.store.Current().Org.SyncedAt, nil),
			))
		} else {
			cmds = append(cmds, m.client.GetIssues(m.store.Current().Org.SyncedAt, nil))

			issues, err := m.store.Issues()
			if err != nil {
				return m, returnError(err)
			}
			m.updateTableRows(issues)
		}

	case client.GetUsersRes:
		if msg.After != nil {
			cmds = append(cmds, m.client.GetUsers(msg.After))
		}
		err := m.store.StoreUsers(msg.Result)
		if err != nil {
			return m, returnError(err)
		}

	case client.GetProjectsRes:
		if msg.After != nil {
			cmds = append(cmds, m.client.GetProjects(msg.After))
		}
		err := m.store.StoreProjects(msg.Result)
		if err != nil {
			return m, returnError(err)
		}

	case client.GetTeamsRes:
		if msg.After != nil {
			cmds = append(cmds, m.client.GetTeams(msg.After))
		}
		err := m.store.StoreTeams(msg.Result)
		if err != nil {
			return m, returnError(err)
		}

	case client.GetIssuesRes:
		if msg.After != nil {
			cmds = append(cmds, m.client.GetIssues(m.store.Current().Org.SyncedAt, msg.After))
		}
		err := m.store.StoreIssues(msg.Result)
		if err != nil {
			return m, returnError(err)
		}

		cmds = append(cmds, m.updateIssues())

		if msg.After == nil {
			if err := m.store.Synced(); err != nil {
				return m, returnError(err)
			}
			m.syncing = false
		}

	case tea.WindowSizeMsg:
		switch m.currView {
		case ViewAll:
			m.table.SetWidth(msg.Width)
		case ViewProject:
			m.prjTable.SetWidth(projectsTableWidth)
			m.table.SetWidth(msg.Width - projectsTableWidth)
		}
		m.table.SetHeight(msg.Height - 4)
		m.prjTable.SetHeight(msg.Height - 5)
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.focus.current() == FocusIssues {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.focus.current() == FocusProjects {
		m.prjTable, cmd = m.prjTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
