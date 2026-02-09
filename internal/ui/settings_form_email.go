package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/basecamp/once/internal/docker"
)

const (
	emailServerField = iota
	emailPortField
	emailUsernameField
	emailPasswordField
	emailFromField
)

type SettingsFormEmail struct {
	settings docker.ApplicationSettings
	form     Form
}

func NewSettingsFormEmail(settings docker.ApplicationSettings) SettingsFormEmail {
	serverField := NewTextField("smtp.example.com")
	serverField.SetValue(settings.SMTP.Server)

	portField := NewTextField("587")
	portField.SetCharLimit(5)
	portField.SetValue(settings.SMTP.Port)

	usernameField := NewTextField("user@example.com")
	usernameField.SetValue(settings.SMTP.Username)

	passwordField := NewTextField("password")
	passwordField.SetEchoPassword()
	passwordField.SetValue(settings.SMTP.Password)

	fromField := NewTextField("noreply@example.com")
	fromField.SetValue(settings.SMTP.From)

	return SettingsFormEmail{
		settings: settings,
		form: NewForm("Done",
			FormItem{Label: "SMTP Server", Field: serverField},
			FormItem{Label: "SMTP Port", Field: portField},
			FormItem{Label: "SMTP Username", Field: usernameField},
			FormItem{Label: "SMTP Password", Field: passwordField},
			FormItem{Label: "SMTP From", Field: fromField},
		),
	}
}

func (m SettingsFormEmail) Title() string {
	return "Email"
}

func (m SettingsFormEmail) Init() tea.Cmd {
	return nil
}

func (m SettingsFormEmail) Update(msg tea.Msg) (SettingsSection, tea.Cmd) {
	var (
		action FormAction
		cmd    tea.Cmd
	)
	m.form, action, cmd = m.form.Update(msg)

	switch action {
	case FormSubmitted:
		m.settings.SMTP.Server = m.form.TextField(emailServerField).Value()
		m.settings.SMTP.Port = m.form.TextField(emailPortField).Value()
		m.settings.SMTP.Username = m.form.TextField(emailUsernameField).Value()
		m.settings.SMTP.Password = m.form.TextField(emailPasswordField).Value()
		m.settings.SMTP.From = m.form.TextField(emailFromField).Value()
		return m, func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	case FormCancelled:
		return m, func() tea.Msg { return SettingsSectionCancelMsg{} }
	}

	return m, cmd
}

func (m SettingsFormEmail) View() string {
	return m.form.View()
}
