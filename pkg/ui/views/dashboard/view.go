package dashboard

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/atoms/hover"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/layouts"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

func ageTextAndColor(issue store.Issue) (string, color.Color) {
	createdAgeDays := -time.Until(issue.CreatedAt).Hours() / 24
	updatedAgeDays := -time.Until(issue.UpdatedAt).Hours() / 24

	baseRed := 120.0
	baseGreen := 120.0
	baseBlue := 120.0

	// Calculate brightness adjustment
	brightnessFactor := -createdAgeDays*(0.04/3.0) + 0.7
	brightnessFactor = math.Max(1, brightnessFactor) // Ensure brightness factor is at least 1

	// Calculate blueness based on how recently the issue was updated
	blueness := 0.0
	if updatedAgeDays <= 20 {
		blueness = (20 - updatedAgeDays) / 20
	}

	// Apply brightness adjustment to all channels
	adjustedRed := baseRed * brightnessFactor
	adjustedGreen := (baseGreen + (255-baseGreen)*blueness*0.3) * brightnessFactor
	adjustedBlue := (baseBlue + (255-baseBlue)*blueness*0.6) * brightnessFactor

	// Ensure the adjusted values do not go below the base color values
	adjustedRed = math.Max(baseRed, math.Min(adjustedRed, 255))       // Clamp between baseRed and 255
	adjustedGreen = math.Max(baseGreen, math.Min(adjustedGreen, 255)) // Clamp between baseGreen and 255
	adjustedBlue = math.Max(baseBlue, math.Min(adjustedBlue, 255))    // Clamp between baseBlue and 255

	ageColorHex := fmt.Sprintf("#%02x%02x%02x", uint8(adjustedRed), uint8(adjustedGreen), uint8(adjustedBlue))

	var ageText string

	age := -time.Until(issue.CreatedAt)
	switch {
	case age < time.Minute:
		ageText = "<1m"
	case age < time.Hour:
		ageText = fmt.Sprintf("%dm", int(age.Minutes()))
	case age < time.Hour*24:
		ageText = fmt.Sprintf("%dh", int(age.Hours()))
	case age < time.Hour*24*7:
		ageText = fmt.Sprintf("%dd", int(age.Hours()/24))
	case age < time.Hour*24*7*30:
		ageText = fmt.Sprintf("%dw", int(age.Hours()/24/7))
	case age < time.Hour*24*7*30*12:
		ageText = fmt.Sprintf("%dM", int(age.Hours()/24/7/30))
	default:
		ageText = fmt.Sprintf("%dy", int(age.Hours()/24/7/30/12))
	}

	return ageText, color.Focusable(ageColorHex, "#888")
}

func (m *Model) updateTableCols() {
	col := color.Simple("#bbb")
	colH := color.Simple("#3f98b5")

	if !m.sortMode {
		colH = col
	}

	cols := []*table.Column{
		table.NewColumn(text.KeymapText("project", col, 0, colH, text.B), 3, table.WithMaxWidth(20)),
		table.NewColumn(text.KeymapText("title", col, 0, colH, text.B), 10, table.WithAutoFill()),
		table.NewColumn(text.Colored("", col), 0, table.WithMaxWidth(4), table.WithMinWidth(4)),
		table.NewColumn(text.KeymapText("assignee", col, 0, colH, text.B), 1, table.WithMaxWidth(10)),
		table.NewColumn(text.KeymapText("state", col, 4, colH, text.B), 1, table.WithMaxWidth(10)),
		table.NewColumn(text.KeymapText("prio", col, 1, colH, text.B), 0.5, table.WithMaxWidth(10)),
		table.NewColumn(text.KeymapText("age", col, 1, colH, text.B), 0.5, table.WithMaxWidth(6)),
		table.NewColumn(text.KeymapText("team", col, 3, colH, text.B), 1.5, table.WithMaxWidth(15)),
		table.NewColumn(text.Colored("labels", col, text.B), 3, table.WithMaxWidth(40)),
	}

	m.table.SetColumns(cols)
}

func (m *Model) renderTopBar() string {
	tinearLogo := text.Chip(
		"Tinear",
		color.Simple("#ccc"),
		color.Simple("#2D4F67"),
		text.Arrow("#527269", false),
		text.B,
	).Focused()

	orgName := text.Chip(
		m.store.Current().Me.Name,
		color.Simple("#000"),
		color.Simple("#527269"),
		text.Arrow("#76946A", false),
	).Focused()

	me := text.Chip(
		m.store.Current().Org.Name,
		color.Simple("#000"),
		color.Simple("#76946A"),
		text.Arrow("", false),
	).Focused()

	var syncedAt string
	if m.syncing {
		syncedAt = text.Colored("syncing...", color.Simple("#444")).Focused()
	} else {
		syncedAt = text.Colored(
			fmt.Sprintf("synced at %s", m.store.Current().SyncedAt.Format(time.DateTime)), color.Simple("#444"),
		).Focused()
	}

	pad := func(s string, p int) string {
		return lipgloss.NewStyle().Padding(0, p).Render(s)
	}

	return pad(layouts.SpaceBetween(
		m.width-2,
		tinearLogo+orgName+me,
		syncedAt,
	), 1)
}

func (m *Model) renderIssues() string {
	return m.table.View()
}

func (m *Model) updateTableRows(issues []store.Issue) {
	rows := make([]*table.Row, 0, len(issues))

	renderPrio := func(p store.Prio, brighten float64) text.Focusable {
		switch p {
		case 1:
			return text.Colored("Urgent", color.Focusable("#e03a43", "#888").Brighten(brighten), text.B)
		case 2:
			return text.Colored("High", color.Focusable("#d47248", "#888").Brighten(brighten))
		case 3:
			return text.Colored("Medium", color.Focusable("#806b38", "#888").Brighten(brighten))
		case 4:
			return text.Colored("Low", color.Focusable("#4a4a4a", "#888").Brighten(brighten))
		}
		return text.Plain("")
	}

	for _, issue := range issues {
		titleNormal := text.Focusable(text.Colored(issue.Title, color.Focusable("#eee", "#888").Darken(0.2)))
		titleSelected := text.Focusable(text.Colored(issue.Title, color.Focusable("#eee", "#888").Brighten(0.2)))

		var projectNormal, projectSelected text.Focusable
		if issue.Project.Name != "" {
			projectNormal = text.Colored(issue.Project.Name, color.Focusable(issue.Project.Color, "#888"))
			projectSelected = text.Colored(issue.Project.Name, color.Focusable(issue.Project.Color, "#888").Brighten(0.2))
		} else {
			projectNormal = text.Colored("", color.Focusable("#555", "#555"))
			projectSelected = text.Colored("", color.Focusable("#555", "#555").Brighten(0.2))
		}

		var assigneeNormal, assigneeSelected text.Focusable
		if issue.Assignee.IsMe {
			assigneeNormal = text.Colored(issue.Assignee.DisplayName, color.Focusable("#76946A", "#888"))
			assigneeSelected = text.Colored(issue.Assignee.DisplayName, color.Focusable("#76946A", "#888").Brighten(0.2))
		} else {
			assigneeNormal = text.Colored(issue.Assignee.DisplayName, color.Focusable("#888", "#888"))
			assigneeSelected = text.Colored(issue.Assignee.DisplayName, color.Focusable("#888", "#888").Brighten(0.2))
		}

		stateNormal := text.Colored(issue.State.Name, color.Focusable(issue.State.Color, "#888"))
		stateSelected := text.Colored(issue.State.Name, color.Focusable(issue.State.Color, "#888").Brighten(0.2))

		priorityNormal := renderPrio(issue.Priority, 0)
		prioritySelected := renderPrio(issue.Priority, 0.2)

		teamNormal := text.Colored(issue.Team.Name, color.Focusable(issue.Team.Color, "#888"))
		teamSelected := text.Colored(issue.Team.Name, color.Focusable(issue.Team.Color, "#888").Brighten(0.2))

		parsedLabels, err := issue.Labels.Parse()
		if err != nil {
			log.Printf("couldn't parse labels: %v\n", err.Error())
		}

		var labelsNormal, labelsSelected []text.Focusable
		for _, label := range parsedLabels {
			labelsNormal = append(labelsNormal, text.Chip(
				label.Name,
				color.Focusable("#eee", "#888"),
				color.Focusable(label.Color, "#444").Darken(0.5),
			))
			labelsSelected = append(labelsSelected, text.Chip(
				label.Name,
				color.Focusable("#eee", "#888").Brighten(0.2),
				color.Focusable(label.Color, "#444").Darken(0.3),
			))
		}

		ageText, ageColor := ageTextAndColor(issue)
		ageNormal := text.Colored(ageText, ageColor)
		ageSelected := text.Colored(ageText, ageColor.Brighten(0.2))

		labelsNormalT := text.Joined("", labelsNormal...)
		labelsSelectedT := text.Joined("", labelsSelected...)

		pinnedText := ""
		if issue.Pinned {
			pinnedText = ""
		}
		pinnedNormal := text.Colored(pinnedText, color.Focusable("#5fa0b8", "#888"), text.B)
		pinnedSelected := text.Colored(pinnedText, color.Focusable("#5fa0b8", "#888").Brighten(0.2), text.B)

		row := &table.Row{
			Identifier: issue.ID,
			Items: []table.RowItem{
				{Normal: projectNormal, Selected: projectSelected},
				{Normal: titleNormal, Selected: titleSelected},
				{Normal: pinnedNormal, Selected: pinnedSelected},
				{Normal: assigneeNormal, Selected: assigneeSelected},
				{Normal: stateNormal, Selected: stateSelected},
				{Normal: priorityNormal, Selected: prioritySelected},
				{Normal: ageNormal, Selected: ageSelected},
				{Normal: teamNormal, Selected: teamSelected},
				{Normal: labelsNormalT, Selected: labelsSelectedT},
			},
		}

		rows = append(rows, row)
	}

	m.table.SetRows(rows)
}

func (m *Model) View() string {
	header := m.renderTopBar()
	issues := m.renderIssues()
	filter := m.input.View()

	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		header,
		"",
		issues,
		"",
		filter,
	)

	if m.hovered == nil {
		return mainContent
	}

	floatingContent := hover.HoverIssue(*m.hovered, m.width-2, m.height-1, true)
	floatingContentHeight := lipgloss.Height(floatingContent)

	issueOffset := m.table.TopOffset() + lipgloss.Height(header) + 3

	// if too close to the bottom
	if floatingContentHeight > m.height-issueOffset-5 {
		issueOffset = max(issueOffset-floatingContentHeight-1, 2)
	}

	return layouts.PlaceOverlay(
		layouts.NewPosition(0, issueOffset),
		floatingContent,
		mainContent,
	)
}
