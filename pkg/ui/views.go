package ui

import (
	"log"

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
    // return tableStyle.Render(m.content.View())
    return m.content.View()
}

func (m *Model) renderDetails() string {
    switch m.currentState {
    case instanceState:
        return utils.PrettyYAML(m.instances[m.cursor])
    case optionState:
        return utils.PrettyYAML(m.options[m.cursor])
    case filesystemState:
        return utils.PrettyYAML(m.filesystems[m.cursor])
    case sshState:
        return utils.PrettyYAML(m.sshKeys[m.cursor])
    default:
        return "Unknown State"
    }
}

func (m *Model) View() string {
    content := m.renderContent()
	return content

    if m.detailsPanel {
        details := m.renderDetails()
        content = lg.JoinHorizontal(lg.Top, content, details)
    }

	log.Println("View: ", m.currentState)
    return lg.JoinVertical(lg.Left,
        m.renderHeader(),
        content,
        m.renderFooter(),
    )
}
