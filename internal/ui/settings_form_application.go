package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/basecamp/once/internal/docker"
)

const (
	appImageField = iota
	appHostnameField
	appTLSField
)

type SettingsFormApplication struct {
	settings docker.ApplicationSettings
	form     Form
}

func NewSettingsFormApplication(settings docker.ApplicationSettings) SettingsFormApplication {
	imageField := NewTextField("user/repo:tag")
	imageField.SetValue(settings.Image)

	hostnameField := NewTextField("app.example.com")
	hostnameField.SetValue(settings.Host)

	tlsField := NewCheckboxField("Enabled", !settings.DisableTLS)
	tlsField.SetDisabledWhen(func() (bool, string) {
		if docker.IsLocalhost(hostnameField.Value()) {
			return true, "Not available for localhost"
		}
		return false, ""
	})

	return SettingsFormApplication{
		settings: settings,
		form: NewForm("Done",
			FormItem{Label: "Image", Field: imageField},
			FormItem{Label: "Hostname", Field: hostnameField},
			FormItem{Label: "TLS", Field: tlsField},
		),
	}
}

func (m SettingsFormApplication) Title() string {
	return "Application"
}

func (m SettingsFormApplication) Init() tea.Cmd {
	return nil
}

func (m SettingsFormApplication) Update(msg tea.Msg) (SettingsSection, tea.Cmd) {
	var (
		action FormAction
		cmd    tea.Cmd
	)
	m.form, action, cmd = m.form.Update(msg)

	switch action {
	case FormSubmitted:
		m.settings.Image = m.form.TextField(appImageField).Value()
		m.settings.Host = m.form.TextField(appHostnameField).Value()
		m.settings.DisableTLS = !m.form.CheckboxField(appTLSField).Checked()
		return m, func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	case FormCancelled:
		return m, func() tea.Msg { return SettingsSectionCancelMsg{} }
	}

	return m, cmd
}

func (m SettingsFormApplication) View() string {
	return m.form.View()
}
