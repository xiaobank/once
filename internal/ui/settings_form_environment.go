package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/docker"
)

type SettingsFormEnvironment struct {
	settings docker.ApplicationSettings
	form     Form
}

func NewSettingsFormEnvironment(settings docker.ApplicationSettings) SettingsFormEnvironment {
	return SettingsFormEnvironment{
		settings: settings,
		form:     NewForm("Done"),
	}
}

func (m SettingsFormEnvironment) Title() string {
	return "Environment"
}

func (m SettingsFormEnvironment) Init() tea.Cmd {
	return nil
}

func (m SettingsFormEnvironment) Update(msg tea.Msg) (SettingsSection, tea.Cmd) {
	var (
		action FormAction
		cmd    tea.Cmd
	)
	m.form, action, cmd = m.form.Update(msg)

	switch action {
	case FormSubmitted, FormCancelled:
		return m, func() tea.Msg { return SettingsSectionCancelMsg{} }
	}

	return m, cmd
}

func (m SettingsFormEnvironment) View() string {
	placeholder := lipgloss.NewStyle().
		Foreground(Colors.Border).
		Italic(true).
		Render("(Environment variable editing coming soon)")

	return lipgloss.JoinVertical(lipgloss.Left,
		placeholder,
		"",
		m.form.View(),
	)
}
