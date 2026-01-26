package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type installKeyMap struct {
	Back key.Binding
}

func (k installKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back}
}

func (k installKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Back}}
}

var installKeys = installKeyMap{
	Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

type Install struct {
	width, height int
	help          help.Model
	form          InstallForm
}

func NewInstall() Install {
	return Install{
		help: help.New(),
		form: NewInstallForm(),
	}
}

func (m Install) Init() tea.Cmd {
	return m.form.Init()
}

func (m Install) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.SetWidth(m.width)
		m.form, _ = m.form.Update(msg)
	case tea.KeyMsg:
		if key.Matches(msg, installKeys.Back) {
			return m, func() tea.Msg { return navigateToDashboardMsg{} }
		}
	case InstallFormCancelMsg:
		return m, func() tea.Msg { return navigateToDashboardMsg{} }
	case InstallFormSubmitMsg:
		// TODO: proceed with installation
		return m, nil
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Install) View() string {
	title := Styles.Title.Width(m.width).Align(lipgloss.Center).Render("Install")

	formView := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(m.form.View())

	helpView := m.help.View(installKeys)
	helpLine := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(helpView)

	topContent := title + "\n\n" + formView
	topHeight := lipgloss.Height(topContent)
	bottomHeight := lipgloss.Height(helpLine)
	middleHeight := max(m.height-topHeight-bottomHeight, 0)

	middle := ""
	for range middleHeight {
		middle += "\n"
	}

	return topContent + middle + helpLine
}
