package ui

import (
	"fmt"
	"lambdactl/pkg/utils"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var viewContent, errorContent, output string

	// Header content
	header := headerStyle.Width(m.width - 2).Render("Lambda Cloud - Running Instances")

	// Footer content
	footer := footerStyle.Width(m.width - 2).Render(fmt.Sprintf("(q) Quit (enter) Select (tab) Switch View (esc) Back - cursor: %d, offset: %d, viewport: %d, length: %d", m.instanceTable.cursor, m.instanceTable.offset, m.instanceTable.viewport, m.instanceTable.length))

	// Dispatch to state's view func
	switch m.currentState {
	case runningState:
		viewContent = runningView(m)
	case optionState:
		viewContent = optionView(m)
	case launchState:
		viewContent = launchView(m)
	default:
		viewContent = runningView(m)
	}

	// Show error message in the footer area if it exists
	if m.errorMsg != "" {
		errorContent = errorStyle.Render("Error: " + m.errorMsg)
		// Join header, main content, and error until it times out
		output = lipgloss.JoinVertical(lipgloss.Top, header, viewContent, errorContent)
	} else {
		// Join header, main content, and footer
		output = lipgloss.JoinVertical(lipgloss.Top, header, viewContent, footer)
	}

	// If in optionState, show the modal on top of the running content
	if m.currentState == detailState {
		// Render the modal but preserve the background content
		modal := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, detailView(m))

		// Overlay the modal on top of the current view
		output = lipgloss.JoinVertical(lipgloss.Top, output, modal)
	}
	return output
}

func runningView(m Model) string {
	return tableStyle.Render(m.instanceTable.View())
}

func detailView(m Model) string {
	return modalStyle.Render(utils.PrettyYAML(m.selectedInstance))
}

func optionView(m Model) string {
	return tableStyle.Render(m.optionTable.View())
}

func launchView(m Model) string {
	modalContent := modalStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			"Launch this option?",
			utils.PrettyYAML(m.selectedOption),
			"\n(q) Quit (esc) Back (l) Launch",
		),
	)

	// Place modal in center of the screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalContent)
}
