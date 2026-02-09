package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basecamp/once/internal/docker"
)

func TestSettingsFormApplication_InitialState_NonLocalhost(t *testing.T) {
	settings := docker.ApplicationSettings{
		Image:      "nginx:latest",
		Host:       "app.example.com",
		DisableTLS: false,
	}
	form := NewSettingsFormApplication(settings)

	assert.Equal(t, 0, form.form.Focused())
	assert.Equal(t, "nginx:latest", form.form.TextField(appImageField).Value())
	assert.Equal(t, "app.example.com", form.form.TextField(appHostnameField).Value())
	assert.True(t, form.form.CheckboxField(appTLSField).Checked())
}

func TestSettingsFormApplication_InitialState_Localhost(t *testing.T) {
	settings := docker.ApplicationSettings{
		Image:      "nginx:latest",
		Host:       "chat.localhost",
		DisableTLS: false,
	}
	form := NewSettingsFormApplication(settings)

	assert.Equal(t, "chat.localhost", form.form.TextField(appHostnameField).Value())
	assert.True(t, form.form.CheckboxField(appTLSField).Checked(), "checkbox is checked (DisableTLS=false)")
}

func TestSettingsFormApplication_TabNavigation(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})
	assert.Equal(t, 0, form.form.Focused())

	form = applicationPressTab(form)
	assert.Equal(t, 1, form.form.Focused(), "hostname")

	form = applicationPressTab(form)
	assert.Equal(t, 2, form.form.Focused(), "tls")

	form = applicationPressTab(form)
	assert.Equal(t, 3, form.form.Focused(), "done button")

	form = applicationPressTab(form)
	assert.Equal(t, 4, form.form.Focused(), "cancel button")

	form = applicationPressTab(form)
	assert.Equal(t, 0, form.form.Focused(), "wraps to image")
}

func TestSettingsFormApplication_ShiftTabNavigation(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})

	form = applicationPressShiftTab(form)
	assert.Equal(t, 4, form.form.Focused(), "cancel button")

	form = applicationPressShiftTab(form)
	assert.Equal(t, 3, form.form.Focused(), "done button")
}

func TestSettingsFormApplication_SpaceTogglesTLS(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})
	assert.True(t, form.form.CheckboxField(appTLSField).Checked())

	form = applicationPressTab(form)
	form = applicationPressTab(form)
	assert.Equal(t, 2, form.form.Focused())

	form = applicationPressSpace(form)
	assert.False(t, form.form.CheckboxField(appTLSField).Checked())

	form = applicationPressSpace(form)
	assert.True(t, form.form.CheckboxField(appTLSField).Checked())
}

func TestSettingsFormApplication_SpaceDoesNotToggleTLSForLocalhost(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "chat.localhost"})

	form = applicationPressTab(form)
	form = applicationPressTab(form)
	assert.Equal(t, 2, form.form.Focused())

	form = applicationPressSpace(form)
	assert.True(t, form.form.CheckboxField(appTLSField).Checked(), "toggle ignored for localhost")
}

func TestSettingsFormApplication_TLSShowsDisabledForLocalhost(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})
	assert.Equal(t, "[x] Enabled", form.form.CheckboxField(appTLSField).View())

	form = applicationPressTab(form)
	form = applicationTypeText(form, ".localhost")
	assert.Equal(t, "Not available for localhost", form.form.CheckboxField(appTLSField).View())

	form = applicationClearAndType(form, "app.example.com")
	assert.Equal(t, "[x] Enabled", form.form.CheckboxField(appTLSField).View())
}

func TestSettingsFormApplication_Submit(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{
		Name:  "myapp",
		Image: "nginx:latest",
		Host:  "app.example.com",
	})

	for range 3 {
		form = applicationPressTab(form)
	}

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	submitMsg, ok := msg.(SettingsSectionSubmitMsg)
	require.True(t, ok, "expected SettingsSectionSubmitMsg, got %T", msg)
	assert.Equal(t, "myapp", submitMsg.Settings.Name)
	assert.Equal(t, "nginx:latest", submitMsg.Settings.Image)
	assert.Equal(t, "app.example.com", submitMsg.Settings.Host)
	assert.False(t, submitMsg.Settings.DisableTLS)
}

func TestSettingsFormApplication_Cancel(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})

	for range 4 {
		form = applicationPressTab(form)
	}

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(SettingsSectionCancelMsg)
	assert.True(t, ok, "expected SettingsSectionCancelMsg, got %T", msg)
}

func TestSettingsFormEmail_InitialState(t *testing.T) {
	settings := docker.ApplicationSettings{
		SMTP: docker.SMTPSettings{
			Server:   "smtp.example.com",
			Port:     "587",
			Username: "user@example.com",
			Password: "secret",
			From:     "noreply@example.com",
		},
	}
	form := NewSettingsFormEmail(settings)

	assert.Equal(t, 0, form.form.Focused())
	assert.Equal(t, "smtp.example.com", form.form.TextField(emailServerField).Value())
	assert.Equal(t, "587", form.form.TextField(emailPortField).Value())
	assert.Equal(t, "user@example.com", form.form.TextField(emailUsernameField).Value())
	assert.Equal(t, "secret", form.form.TextField(emailPasswordField).Value())
	assert.Equal(t, "noreply@example.com", form.form.TextField(emailFromField).Value())
}

func TestSettingsFormEmail_TabNavigation(t *testing.T) {
	form := NewSettingsFormEmail(docker.ApplicationSettings{})
	assert.Equal(t, 0, form.form.Focused())

	form = emailPressTab(form)
	assert.Equal(t, 1, form.form.Focused(), "port")

	form = emailPressTab(form)
	assert.Equal(t, 2, form.form.Focused(), "username")

	form = emailPressTab(form)
	assert.Equal(t, 3, form.form.Focused(), "password")

	form = emailPressTab(form)
	assert.Equal(t, 4, form.form.Focused(), "from")

	form = emailPressTab(form)
	assert.Equal(t, 5, form.form.Focused(), "done button")

	form = emailPressTab(form)
	assert.Equal(t, 6, form.form.Focused(), "cancel button")

	form = emailPressTab(form)
	assert.Equal(t, 0, form.form.Focused(), "wraps to server")
}

func TestSettingsFormEmail_Submit(t *testing.T) {
	settings := docker.ApplicationSettings{
		Name: "myapp",
		SMTP: docker.SMTPSettings{
			Server: "smtp.example.com",
			Port:   "587",
		},
	}
	form := NewSettingsFormEmail(settings)

	for range 5 {
		form = emailPressTab(form)
	}

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	submitMsg, ok := msg.(SettingsSectionSubmitMsg)
	require.True(t, ok, "expected SettingsSectionSubmitMsg, got %T", msg)
	assert.Equal(t, "myapp", submitMsg.Settings.Name)
	assert.Equal(t, "smtp.example.com", submitMsg.Settings.SMTP.Server)
	assert.Equal(t, "587", submitMsg.Settings.SMTP.Port)
}

func TestSettingsFormEmail_Cancel(t *testing.T) {
	form := NewSettingsFormEmail(docker.ApplicationSettings{})

	for range 6 {
		form = emailPressTab(form)
	}

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(SettingsSectionCancelMsg)
	assert.True(t, ok, "expected SettingsSectionCancelMsg, got %T", msg)
}

func TestSettingsFormResources_InitialState(t *testing.T) {
	settings := docker.ApplicationSettings{
		Resources: docker.ContainerResources{
			CPUs:     2,
			MemoryMB: 512,
		},
	}
	form := NewSettingsFormResources(settings)

	assert.Equal(t, 0, form.form.Focused())
	assert.Equal(t, "2", form.form.TextField(resourcesCPUField).Value())
	assert.Equal(t, "512", form.form.TextField(resourcesMemoryField).Value())
}

func TestSettingsFormResources_InitialState_ZeroValues(t *testing.T) {
	form := NewSettingsFormResources(docker.ApplicationSettings{})

	assert.Equal(t, 0, form.form.Focused())
	assert.Equal(t, "", form.form.TextField(resourcesCPUField).Value())
	assert.Equal(t, "", form.form.TextField(resourcesMemoryField).Value())
}

func TestSettingsFormResources_TabNavigation(t *testing.T) {
	form := NewSettingsFormResources(docker.ApplicationSettings{})
	assert.Equal(t, 0, form.form.Focused())

	form = resourcesPressTab(form)
	assert.Equal(t, 1, form.form.Focused(), "memory")

	form = resourcesPressTab(form)
	assert.Equal(t, 2, form.form.Focused(), "done button")

	form = resourcesPressTab(form)
	assert.Equal(t, 3, form.form.Focused(), "cancel button")

	form = resourcesPressTab(form)
	assert.Equal(t, 0, form.form.Focused(), "wraps to cpu")
}

func TestSettingsFormResources_Submit(t *testing.T) {
	form := NewSettingsFormResources(docker.ApplicationSettings{Name: "myapp"})

	form = resourcesTypeText(form, "2")
	form = resourcesPressTab(form)
	form = resourcesTypeText(form, "1024")
	form = resourcesPressTab(form)

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	submitMsg, ok := msg.(SettingsSectionSubmitMsg)
	require.True(t, ok, "expected SettingsSectionSubmitMsg, got %T", msg)
	assert.Equal(t, "myapp", submitMsg.Settings.Name)
	assert.Equal(t, 2, submitMsg.Settings.Resources.CPUs)
	assert.Equal(t, 1024, submitMsg.Settings.Resources.MemoryMB)
}

func TestSettingsFormResources_SubmitBlank(t *testing.T) {
	form := NewSettingsFormResources(docker.ApplicationSettings{})

	form = resourcesPressTab(form)
	form = resourcesPressTab(form)

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	submitMsg, ok := msg.(SettingsSectionSubmitMsg)
	require.True(t, ok, "expected SettingsSectionSubmitMsg, got %T", msg)
	assert.Equal(t, 0, submitMsg.Settings.Resources.CPUs)
	assert.Equal(t, 0, submitMsg.Settings.Resources.MemoryMB)
}

func TestSettingsFormResources_Cancel(t *testing.T) {
	form := NewSettingsFormResources(docker.ApplicationSettings{})

	for range 3 {
		form = resourcesPressTab(form)
	}

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(SettingsSectionCancelMsg)
	assert.True(t, ok, "expected SettingsSectionCancelMsg, got %T", msg)
}

// Helpers

func applicationTypeText(form SettingsFormApplication, text string) SettingsFormApplication {
	for _, r := range text {
		section, _ := form.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		form = section.(SettingsFormApplication)
	}
	return form
}

func applicationClearAndType(form SettingsFormApplication, text string) SettingsFormApplication {
	form.form.TextField(appHostnameField).SetValue("")
	return applicationTypeText(form, text)
}

func applicationPressTab(form SettingsFormApplication) SettingsFormApplication {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return section.(SettingsFormApplication)
}

func applicationPressShiftTab(form SettingsFormApplication) SettingsFormApplication {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	return section.(SettingsFormApplication)
}

func applicationPressSpace(form SettingsFormApplication) SettingsFormApplication {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	return section.(SettingsFormApplication)
}

func emailPressTab(form SettingsFormEmail) SettingsFormEmail {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return section.(SettingsFormEmail)
}

func resourcesPressTab(form SettingsFormResources) SettingsFormResources {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return section.(SettingsFormResources)
}

func resourcesTypeText(form SettingsFormResources, text string) SettingsFormResources {
	for _, r := range text {
		section, _ := form.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		form = section.(SettingsFormResources)
	}
	return form
}
