package dashboard

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/linear/models"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/input"
)

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
			m.updateTables(),
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
			m.input.Blur()
			return m.updateTables()
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

		return tea.Batch(cmd, m.updateTables(withDebounce()))
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

	err := m.store.SetBookmark(selectedIssues...)
	if err != nil {
		return returnError(err)
	}

	m.table.SetVisualMode(false)

	var selected string
	if len(selectedIssues) == 1 {
		selected = selectedIssues[0]
	}

	return m.updateTables(withSelectedIssue(selected))
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

func (m *Model) handleFocus(key tea.KeyMsg) tea.Cmd {
	if key.Type != tea.KeyEscape {
		return nil
	}

	// NOTE: special filter escape layer
	if m.input.Value() != "" {
		m.input.SetValue("")
		if m.focus.current() != FocusFilter {
			return m.updateTables()
		}
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
	onPop := tea.Batch(shiftFocus, m.updateTables())

	if m.focus.push(FocusIssues, onPop) {
		m.prjTable.Blur()
		m.table.Focus()
	}

	return m.updateTables()
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
			return m.updateTables(withCursorAtIssue(0))
		})

		issue, err := m.store.Issue(m.table.SelectedRow())
		if err != nil {
			return returnError(err)
		}

		if issue.Project.ID != "" {
			m.store.SetProject(&issue.Project)
			return m.updateTables(withSelectedProject(issue.Project.ID), withSelectedIssue(issue.ID))
		}

		return m.updateTables(withCursorAtIssue(0))
	case ViewProject:
		m.currView = ViewAll
		m.focus = []focusStackItem{{mode: FocusIssues}}

		m.store.SetProject(nil)
		m.table.SetWidth(m.width)
		m.table.Focus()
		m.prjTable.Blur()
		m.prjTable.SetOnMove(nil)

		return m.updateTables(withSelectedIssue(m.table.SelectedRow()))
	}

	return nil
}

func (m *Model) handleSelector(key tea.KeyMsg) (cmd tea.Cmd) {
	switch m.focus.current() {
	case FocusIssues:
		onPop := func() tea.Msg {
			m.table.Focus()
			m.updateTableCols()
			return forceUpdate()
		}

		if key.String() != "e" {
			return nil
		}

		if m.focus.push(FocusSelectorPre, onPop) {
			m.table.Blur()
			m.updateTableCols()
		}

	case FocusSelectorPre:
		var suggestion []input.Suggestion
		var mode selectorMode

		onPop := func() tea.Msg {
			m.table.Focus()
			m.selector.Reset()
			return nil
		}

		switch key.String() {
		default:
			return nil
		case "p": // projects
			mode = SelectorModeProject

			projects, err := m.store.Projects()
			if err != nil {
				return returnError(err)
			}

			for _, project := range projects {
				suggestion = append(suggestion, input.Suggestion{
					Identifier: project.ID,
					Title:      project.Name,
					Color:      project.Color,
				})
			}
		case "a": // assignee
			mode = SelectorModeAssignee

			users, err := m.store.Users()
			if err != nil {
				return returnError(err)
			}

			for _, user := range users {
				suggestion = append(suggestion, input.Suggestion{
					Identifier: user.ID,
					Title:      user.DisplayName,
				})
			}
		case "e": // state
			mode = SelectorModeState

			issues, err := m.store.Issues(m.table.SelectedRows()...)
			if err != nil {
				return returnError(err)
			}

			var teamID string

			for _, issue := range issues {
				if teamID == "" {
					teamID = issue.Team.ID
				} else if teamID != issue.Team.ID {
					// means issues selected go across two teams
					return returnError(errors.New("issues selected are not of the same team"))
				}
			}

			states, err := m.store.States(teamID)
			if err != nil {
				return returnError(err)
			}

			for _, state := range states {
				suggestion = append(suggestion, input.Suggestion{
					Identifier: state.ID,
					Title:      state.Name,
					Color:      state.Color,
				})
			}
		case "r": // prio
			mode = SelectorModePriority

			suggestion = []input.Suggestion{
				{Identifier: "0", Title: "No Priority", Color: "#555555"},
				{Identifier: "1", Title: "Urgent", Color: "#e03a43"},
				{Identifier: "2", Title: "High", Color: "#d47248"},
				{Identifier: "3", Title: "Medium", Color: "#806b38"},
				{Identifier: "4", Title: "Low", Color: "#4a4a4a"},
			}
		case "m": // team
			mode = SelectorModeTeam

			teams, err := m.store.Teams()
			if err != nil {
				return returnError(err)
			}

			for _, team := range teams {
				suggestion = append(suggestion, input.Suggestion{
					Identifier: team.ID,
					Title:      team.Name,
					Color:      team.Color,
				})
			}
		case "l": // labels
			mode = SelectorModeLabels

			issues, err := m.store.Issues(m.table.SelectedRows()...)
			if err != nil {
				return returnError(err)
			}

			var teamID string

			for _, issue := range issues {
				if teamID == "" {
					teamID = issue.Team.ID
					continue
				}
				if teamID != issue.Team.ID {
					// just set it only default labels
					teamID = ""
					break
				}
			}

			labels, err := m.store.Labels(teamID)
			if err != nil {
				return returnError(err)
			}

			for _, label := range labels {
				suggestion = append(suggestion, input.Suggestion{
					Identifier: label.ID,
					Title:      label.Name,
					Color:      label.Color,
				})
			}

		case "t": // title
			mode = SelectorModeTitle

			issue, err := m.store.Issue(m.table.SelectedRow())
			if err != nil {
				return returnError(err)
			}

			m.selector.SetValue(issue.Title)
		}

		m.focus.pop()()

		if m.focus.push(FocusSelector, tea.Batch(onPop, m.updateTables())) {
			m.table.Blur()
			m.selector.SetSuggestions(suggestion)
			m.selectorMode = mode
		}
	case FocusSelector:
		m.selector, cmd = m.selector.Update(key)

		if key.Type != tea.KeyEnter {
			return cmd
		}

		m.table.SetVisualMode(false)

		selectedIssueIDs := m.table.SelectedRows()
		selectedIssues, err := m.store.Issues(selectedIssueIDs...)
		if err != nil {
			return returnError(err)
		}

		suggested := m.selector.Highlighted()

		if m.selectorMode != SelectorModeTitle && suggested == nil {
			return nil
		}

		var updatedField store.UpdateIssueField
		var updatedOpt client.IssueUpdateOpt
		var updatedValue any

		switch m.selectorMode {
		default:
			return nil

		case SelectorModeAssignee:
			updatedField = store.UpdateIssueFieldAssignee
			updatedOpt = client.WithSetAssignee(suggested.Identifier)
			updatedValue = suggested.Identifier

		case SelectorModePriority:
			updatedField = store.UpdateIssueFieldPrio
			prio, err := strconv.ParseInt(suggested.Identifier, 10, 64)
			if err != nil {
				return returnError(err)
			}
			updatedOpt = client.WithSetPrio(prio)
			updatedValue = suggested.Identifier

		case SelectorModeProject:
			updatedField = store.UpdateIssueFieldProject
			updatedOpt = client.WithSetProject(suggested.Identifier)
			if suggested.Title == "(No Project)" {
				updatedOpt = client.WithSetProject(models.NullString)
			}
			updatedValue = suggested.Identifier

		case SelectorModeTeam:
			updatedField = store.UpdateIssueFieldTeam
			updatedOpt = client.WithSetTeam(suggested.Identifier)
			updatedValue = suggested.Identifier

		case SelectorModeState:
			updatedField = store.UpdateIssueFieldState
			updatedOpt = client.WithSetState(suggested.Identifier)
			updatedValue = suggested.Identifier

		case SelectorModeTitle:
			updatedField = store.UpdateIssueFieldTitle
			inputValue := m.selector.Value()
			if inputValue == "" {
				return nil
			}
			updatedOpt = client.WithSetTitle(inputValue)
			updatedValue = inputValue

		case SelectorModeLabels:
			updatedOpt = client.WithAddLabels(suggested.Identifier)
			updatedValue = suggested.Identifier
		}

		// err = m.store.UpdateIssues(updatedField, updatedValue, selectedIssueIDs...)
		// if err != nil {
		// 	return returnError(err)
		// }
		_ = updatedField
		_ = updatedValue

		onFail := func() tea.Msg {
			err := m.store.StoreIssues(selectedIssues)
			if err != nil {
				return returnError(err)
			}
			return m.updateTables()
		}

		updateIssues := m.client.UpdateIssues(selectedIssueIDs, onFail, updatedOpt)

		return tea.Batch(m.focus.pop(), updateIssues)
	}
	return nil
}

func (m *Model) handleDebug(key tea.KeyMsg) tea.Cmd {
	if key.String() == "+" {
		m.debug = fmt.Sprintf(`focus=%v`, m.focus.current())
	}
	if key.String() == "-" {
		m.debug = ""
	}
	return nil
}

type (
	updateTablesMsg struct {
		issues        []store.Issue
		projects      []store.Project
		issue         string
		project       string
		issueCursorAt int
	}
	updateTablesOpt struct {
		issue         string
		project       string
		debounce      bool
		updateColumns bool
		cursorAt      int
	}
	updateTablesOptFunc func(opt *updateTablesOpt)
)

func withSelectedIssue(selected string) updateTablesOptFunc {
	return func(opt *updateTablesOpt) {
		opt.issue = selected
	}
}

func withSelectedProject(selected string) updateTablesOptFunc {
	return func(opt *updateTablesOpt) {
		opt.project = selected
	}
}

func withDebounce() updateTablesOptFunc {
	return func(opt *updateTablesOpt) {
		opt.debounce = true
	}
}

func withCursorAtIssue(n int) updateTablesOptFunc {
	return func(opt *updateTablesOpt) {
		opt.cursorAt = n
	}
}

func (m *Model) updateTables(opts ...updateTablesOptFunc) tea.Cmd {
	var options updateTablesOpt

	options.cursorAt = -1

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

		projects, err := m.store.Projects()
		if err != nil {
			return err
		}

		return updateTablesMsg{
			issues:        issues,
			projects:      projects,
			issue:         options.issue,
			project:       options.project,
			issueCursorAt: options.cursorAt,
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	default:
		log.Printf("received a msg that is not recognized here, type: %T, msg: %v", msg, msg)

	case error:
		m.err = msg
		return m, nil

	case tea.Cmd:
		cmds = append(cmds, msg)

	case tea.KeyMsg:
		cmds = append(cmds, m.handleFilter(msg))
		cmds = append(cmds, m.handleBookmark(msg))
		cmds = append(cmds, m.handleHover(msg))
		cmds = append(cmds, m.handleOpen(msg))
		cmds = append(cmds, m.handleSelector(msg))
		cmds = append(cmds, m.handleSortMode(msg))
		cmds = append(cmds, m.handleClose(msg))
		cmds = append(cmds, m.handleFocus(msg))
		cmds = append(cmds, m.handleProjectSelection(msg))
		cmds = append(cmds, m.handleViews(msg))
		cmds = append(cmds, m.handleDebug(msg))

	case updateTablesMsg:
		m.table.SetLoading(false)
		m.updateTableCols()
		m.updateTableRows(msg.issues)
		m.updateProjectsTable(msg.projects)
		if msg.issue != "" {
			m.table.SetSelectedRow(msg.issue)
		}
		if msg.project != "" {
			m.prjTable.SetSelectedRow(msg.project)
		}
		if msg.issueCursorAt != -1 {
			m.table.SetCursor(msg.issueCursorAt)
		}

	case client.UpdateIssuesResponse:
		if !msg.Success {
			cmds = append(cmds, msg.OnFailCommand)
		}

	case client.GetMeRes:
		orgChanged, err := m.store.StoreOrg(msg.Result.Org)
		if err != nil {
			return m, returnError(err)
		}

		teamsChanged, err := m.store.StoreTeams(msg.Result.Teams)
		if err != nil {
			return m, returnError(err)
		}

		err = m.store.StoreLabels(msg.Result.Labels)
		if err != nil {
			return m, returnError(err)
		}

		err = m.store.StoreStates(msg.Result.States)
		if err != nil {
			return m, returnError(err)
		}

		var teamIDs []string
		for _, team := range msg.Result.Teams {
			teamIDs = append(teamIDs, team.ID)
		}

		if orgChanged || teamsChanged {
			m.table.SetLoading(orgChanged)
			m.syncing = true

			lastSync := m.store.Current().Org.SyncedAt
			if teamsChanged {
				lastSync = time.Now().Add(-6 * 30 * 24 * time.Hour)
			}

			cmds = append(cmds, tea.Batch(
				m.client.GetProjects(nil),
				m.client.GetUsers(nil),
				m.client.GetIssues(lastSync, teamIDs, nil),
			))
		} else {
			cmds = append(cmds, m.client.GetIssues(m.store.Current().Org.SyncedAt, teamIDs, nil))

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

	case client.GetIssuesRes:
		teams, err := m.store.Teams()
		if err != nil {
			return m, returnError(err)
		}

		var teamIDs []string
		for _, team := range teams {
			teamIDs = append(teamIDs, team.ID)
		}

		if msg.After != nil {
			cmds = append(cmds, m.client.GetIssues(m.store.Current().Org.SyncedAt, teamIDs, msg.After))
		}
		err = m.store.StoreIssues(msg.Result)
		if err != nil {
			return m, returnError(err)
		}

		cmds = append(cmds, m.updateTables())

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
