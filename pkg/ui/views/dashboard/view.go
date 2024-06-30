package dashboard

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/store"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

func renderIssues(issues []store.Issue, focused bool) []*table.Row {
	rows := make([]*table.Row, 0)

	for _, issue := range issues {
		titleNormal := text.Colored(issue.Title, color.Focusable("#eee", "#888").Darken(0.2))
		titleSelected := text.Colored(issue.Title, color.Focusable("#eee", "#888").Brighten(0.2))

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
			assigneeNormal = text.Colored("me", color.Focusable("#76946A", "#888"))
			assigneeSelected = text.Colored("me", color.Focusable("#76946A", "#888").Brighten(0.2))
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

		var labelsNormal, labelsSelected []string
		parsedLabels, err := issue.Labels.Parse()
		if err != nil {
			panic(err)
		}
		for _, label := range parsedLabels {
			normal := text.Chip(label.Name, color.Focusable("#eee", "#888"), color.Simple(label.Color).Darken(0.5))
			selected := text.Chip(label.Name, color.Focusable("#eee", "#888").Brighten(0.2), color.Simple(label.Color).Darken(0.3))

			if focused {
				labelsNormal = append(labelsNormal, normal.Focused())
				labelsSelected = append(labelsSelected, selected.Focused())
			} else {
				labelsNormal = append(labelsNormal, normal.Blurred())
				labelsSelected = append(labelsSelected, selected.Blurred())
			}
		}

		labelsNormalT := text.Plain(strings.Join(labelsNormal, ""))
		labelsSelectedT := text.Plain(strings.Join(labelsSelected, ""))

		row := &table.Row{
			Identifier: issue.ID,
			Items: []table.RowItem{
				{
					Normal:   projectNormal,
					Selected: projectSelected,
				},
				{
					Normal:   titleNormal,
					Selected: titleSelected,
				},
				{
					Normal:   assigneeNormal,
					Selected: assigneeSelected,
				},
				{
					Normal:   stateNormal,
					Selected: stateSelected,
				},
				{
					Normal:   priorityNormal,
					Selected: prioritySelected,
				},
				{
					Normal:   teamNormal,
					Selected: teamSelected,
				},
				{
					Normal:   labelsNormalT,
					Selected: labelsSelectedT,
				},
			},
		}

		rows = append(rows, row)
	}

	return rows
}

func renderPrio(p store.Prio, brighten float64) text.Focusable {
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

func (m *Model) renderTableCols(helpMode bool) {
	colColor := color.Simple("#bbb")
	colHColor := color.Simple("#DCA561")

	if !helpMode {
		colHColor = colColor
	}

	cols := []*table.Column{
		table.NewColumn(
			text.KeymapText("project", colColor, 0, colHColor),
			3, table.WithMaxWidth(20),
		),
		table.NewColumn(
			text.KeymapText("title", colColor, 0, colHColor),
			10, table.WithAutoFill(),
		),
		table.NewColumn(
			text.KeymapText("assignee", colColor, 0, colHColor),
			1, table.WithMaxWidth(10),
		),
		table.NewColumn(
			text.KeymapText("state", colColor, 0, colHColor),
			1, table.WithMaxWidth(10),
		),
		table.NewColumn(
			text.KeymapText("prio", colColor, 0, colHColor),
			0.5, table.WithMaxWidth(10),
		),
		table.NewColumn(
			text.KeymapText("team", colColor, 3, colHColor),
			1.5, table.WithMaxWidth(15),
		),
		table.NewColumn(
			text.KeymapText("labels", colColor, 0, colHColor),
			3, table.WithMaxWidth(40),
		),
	}

	m.table.SetColumns(cols)
}

func (m *Model) renderTopBar() string {
	return ""
	// tinearLogo := text.Chip(
	// 	"tinear",
	// 	color.Simple("#ccc"),
	// 	color.Simple("#2D4F67"),
	// 	text.B,
	// )
	//
	// in := text.Colored(
	// 	" in ",
	// 	color.Simple("#555"),
	// )
	//
	// org := text.Colored(
	// 	m.state.Me.OrgName,
	// 	color.Simple("#445f70"),
	// )
	//
	// me := text.Colored(
	// 	m.Me.DisplayName,
	// 	color.Simple("#76946A"),
	// )
	//
	// return layouts.SpaceBetween(
	// 	m.common.Size.Width()-2,
	// 	" "+tinearLogo.Focused(),
	// 	lipgloss.JoinHorizontal(
	// 		lipgloss.Top,
	// 		me.Focused(),
	// 		in.Focused(),
	// 		org.Focused(),
	// 	),
	// )
}

func (m *Model) renderIssues() string {
	return m.table.View()
}

func (m *Model) renderTableRows(issues []store.Issue) {
	rows := renderIssues(issues, m.table.Focused())

	m.table.SetRows(rows)
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		m.renderTopBar(),
		"",
		m.renderIssues(),
	)
}
