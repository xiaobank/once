package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallForm_FillAndSubmit(t *testing.T) {
	form := NewInstallForm()

	form = installTypeText(form, "nginx:latest")
	assert.Equal(t, "nginx:latest", form.ImageRef())

	form = installPressEnter(form)
	assert.Equal(t, 1, form.form.Focused())

	form = installTypeText(form, "myapp.example.com")
	assert.Equal(t, "myapp.example.com", form.Hostname())

	form = installPressEnter(form)
	assert.Equal(t, 2, form.form.Focused(), "submit button")

	form, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	submitMsg, ok := msg.(InstallFormSubmitMsg)
	require.True(t, ok, "expected InstallFormSubmitMsg, got %T", msg)
	assert.Equal(t, "nginx:latest", submitMsg.ImageRef)
	assert.Equal(t, "myapp.example.com", submitMsg.Hostname)
}

func TestInstallForm_TabNavigation(t *testing.T) {
	form := NewInstallForm()
	assert.Equal(t, 0, form.form.Focused())

	form = installPressTab(form)
	assert.Equal(t, 1, form.form.Focused())

	form = installPressTab(form)
	assert.Equal(t, 2, form.form.Focused(), "submit button")

	form = installPressTab(form)
	assert.Equal(t, 3, form.form.Focused(), "cancel button")

	form = installPressTab(form)
	assert.Equal(t, 0, form.form.Focused(), "wraps to first field")
}

func TestInstallForm_ShiftTabNavigation(t *testing.T) {
	form := NewInstallForm()

	form = installPressShiftTab(form)
	assert.Equal(t, 3, form.form.Focused(), "cancel button")

	form = installPressShiftTab(form)
	assert.Equal(t, 2, form.form.Focused(), "submit button")
}

func TestInstallForm_CancelButton(t *testing.T) {
	form := NewInstallForm()

	for range 3 {
		form = installPressTab(form)
	}
	assert.Equal(t, 3, form.form.Focused(), "cancel button")

	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(InstallFormCancelMsg)
	assert.True(t, ok, "expected InstallFormCancelMsg, got %T", msg)
}

// Helpers

func installTypeText(form InstallForm, text string) InstallForm {
	for _, r := range text {
		form, _ = form.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return form
}

func installPressEnter(form InstallForm) InstallForm {
	form, _ = form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return form
}

func installPressTab(form InstallForm) InstallForm {
	form, _ = form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return form
}

func installPressShiftTab(form InstallForm) InstallForm {
	form, _ = form.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	return form
}
