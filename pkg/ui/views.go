package ui

import (
	"lambdactl/pkg/utils"

	lg "github.com/charmbracelet/lipgloss"
)

func (m *Model) renderHeader() string {
	return headerStyle.Render(m.header)
}

func (m *Model) renderFooter() string {
	return footerStyle.Render(m.footer)
}

func (m *Model) renderContent() string {
	return m.content.View()
}

func (m *Model) renderDetails() string {
	switch m.currentState {
	case instanceState:
		return utils.PrettyYAML(m.instances[m.content.Cursor()])
	case optionState:
		return utils.PrettyYAML(m.options[m.content.Cursor()])
	case filesystemState:
		return utils.PrettyYAML(m.filesystems[m.content.Cursor()])
	case sshState:
		return utils.PrettyYAML(m.sshKeys[m.content.Cursor()])
	default:
		return "Unknown State"
	}
}

func (m *Model) View() string {
	content := m.renderContent()

	if m.detailsPanel {
		details := m.renderDetails()
		modalStyle := lg.NewStyle().
			Border(lg.RoundedBorder()).
			Padding(1, 2).
			Background(lg.Color("#333")).
			Foreground(lg.Color("#FFF")).
			Width(50).
			Height(20) // You can adjust the modal size as needed

		modal := modalStyle.Render(details)

		// Center the modal over the content
		modal = lg.Place(m.width, m.height, lg.Center, lg.Center, modal)

		content = lg.JoinVertical(lg.Left,
			m.renderHeader(),
			content,
			modal, // Render details on top
			m.renderFooter(),
		)
		return content
	}

	return lg.JoinVertical(lg.Left,
		m.renderHeader(),
		content,
		m.renderFooter(),
	)
}
