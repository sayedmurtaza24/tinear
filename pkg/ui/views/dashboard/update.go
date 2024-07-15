package dashboard

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sayedmurtaza24/tinear/pkg/client"
	"github.com/sayedmurtaza24/tinear/pkg/store"
)

func returnError(err error) tea.Cmd {
	return func() tea.Msg { return err }
}

func (m *Model) handleSortMode(key tea.KeyMsg) tea.Cmd {
	if m.filterMode {
		return nil
	}

	if !m.sortMode {
		if key.String() == "s" {
			m.sortMode = true
			m.table.Blur()
			m.updateTableCols()
		}
		return nil
	}

	if key.Type == tea.KeyEscape {
		m.sortMode = false
		m.table.Focus()
		m.updateTableCols()

		return m.updateIssues()
	}

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

	m.sortMode = false
	m.table.Focus()

	m.updateTableCols()

	return m.updateIssues()
}

func (m *Model) handleFilter(key tea.KeyMsg) (cmd tea.Cmd) {
	if m.sortMode {
		return nil
	}

	if !m.filterMode {
		if key.Type == tea.KeyEsc && len(m.input.Value()) > 0 && !m.table.VisualMode() {
			m.input.SetValue("")

			return tea.Batch(
				cmd,
				m.updateIssues(),
			)
		}

		if key.String() == "/" {
			m.filterMode = true
			m.input.Focus()
			m.input.SetValue("")

			m.input, cmd = m.input.Update(key)

			return cmd
		}

		return nil
	}

	backspaced := (key.Type == tea.KeyBackspace && m.input.Value() == "/")
	escaped := key.Type == tea.KeyEscape
	applied := key.Type == tea.KeyEnter

	if escaped || backspaced || (applied && m.input.Value() == "/") {
		m.input.SetValue("")
	}
	if backspaced || escaped || applied {
		m.filterMode = false
		m.input.Blur()
		return m.updateIssues()
	}

	m.input, cmd = m.input.Update(key)

	return tea.Batch(
		cmd,
		m.updateIssues(),
	)
}

func (m *Model) handleBookmark(key tea.KeyMsg) tea.Cmd {
	if m.sortMode || m.filterMode {
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

	return m.updateIssues()
}

func (m *Model) handleOpen(key tea.KeyMsg) tea.Cmd {
	if m.sortMode || m.filterMode || m.table.VisualMode() {
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

	url := fmt.Sprintf("%s/%s/issue/%s", linearBaseURL, urlKey, issue.Identifier)

	open(url)

	return nil
}

func (m *Model) handleHover(key tea.KeyMsg) tea.Cmd {
	if key.String() != "K" {
		m.hovered = nil
		if !m.table.Focused() {
			m.table.Focus()
		}
		return nil
	}

	if m.filterMode || m.sortMode {
		return nil
	}

	issue, err := m.store.Issue(m.table.SelectedRow())
	if err != nil {
		return returnError(err)
	}

	m.hovered = issue
	m.table.Blur()

	return nil
}

type updateIssuesMsg []store.Issue

func (m *Model) updateIssues() tea.Cmd {
	return func() tea.Msg {
		m.table.SetLoading(false)

		var issues []store.Issue
		var err error

		currentInput := m.input.Value()

		// debounce
		time.Sleep(150 * time.Millisecond)

		if currentInput != m.input.Value() {
			return nil
		}

		if len(m.input.Value()) > 3 {
			issues, err = m.store.SearchIssues(m.input.Value()[1:])
			if err != nil {
				return err
			}
		} else {
			issues, err = m.store.Issues()
			if err != nil {
				return err
			}
		}

		return updateIssuesMsg(issues)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case error:
		log.Println(msg)
		return m, tea.Quit

	case tea.KeyMsg:
		cmds = append(cmds, m.handleSortMode(msg))
		cmds = append(cmds, m.handleFilter(msg))
		cmds = append(cmds, m.handleBookmark(msg))
		cmds = append(cmds, m.handleHover(msg))
		cmds = append(cmds, m.handleOpen(msg))

	case updateIssuesMsg:
		m.updateTableRows(msg)

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
				m.client.GetIssues(m.store.Current().SyncedAt, nil),
			))
		} else {
			cmds = append(cmds, m.client.GetIssues(m.store.Current().SyncedAt, nil))

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
			cmds = append(cmds, m.client.GetIssues(m.store.Current().SyncedAt, msg.After))
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
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 5)
		m.width = msg.Width
		m.height = msg.Height
	}

	if !m.filterMode {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
