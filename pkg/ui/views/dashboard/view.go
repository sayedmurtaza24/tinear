package dashboard

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sayedmurtaza24/tinear/pkg/linear/issue"
	"github.com/sayedmurtaza24/tinear/pkg/ui/color"
	"github.com/sayedmurtaza24/tinear/pkg/ui/layouts"
	"github.com/sayedmurtaza24/tinear/pkg/ui/molecules/table"
	"github.com/sayedmurtaza24/tinear/pkg/ui/text"
)

func (m *Model) renderTableCols() {
	colColor := color.Simple("#bbb")
	colHColor := color.Simple("#DCA561")

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
	tinearLogo := text.Chip(
		"tinear",
		color.Simple("#ccc"),
		color.Simple("#2D4F67"),
		text.B,
	)

	in := text.Colored(
		" in ",
		color.Simple("#555"),
	)

	org := text.Colored(
		m.state.OrganizationName,
		color.Simple("#445f70"),
	)

	me := text.Colored(
		m.state.Me.DisplayName,
		color.Simple("#76946A"),
	)

	return layouts.SpaceBetween(
		m.common.Size.Width()-2,
		" "+tinearLogo.Focused(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			me.Focused(),
			in.Focused(),
			org.Focused(),
		),
	)
}

func (m *Model) renderIssues() string {
	return m.table.View()
}

func (m *Model) renderTableRows(issues []issue.Issue) {
	rows := issue.IssuesToRows(issues, m.table.Focused())

	m.table.SetRows(rows)
	m.table.SetWidth(m.common.Size.Width())
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
