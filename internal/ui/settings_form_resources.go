package ui

import (
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/basecamp/once/internal/docker"
)

const (
	resourcesCPUField = iota
	resourcesMemoryField
)

type SettingsFormResources struct {
	settings docker.ApplicationSettings
	form     Form
}

func NewSettingsFormResources(settings docker.ApplicationSettings) SettingsFormResources {
	cpuField := NewTextField("e.g. 2")
	cpuField.SetCharLimit(10)
	cpuField.SetDigitsOnly(true)
	if settings.Resources.CPUs != 0 {
		cpuField.SetValue(strconv.Itoa(settings.Resources.CPUs))
	}

	memoryField := NewTextField("e.g. 512")
	memoryField.SetCharLimit(10)
	memoryField.SetDigitsOnly(true)
	if settings.Resources.MemoryMB != 0 {
		memoryField.SetValue(strconv.Itoa(settings.Resources.MemoryMB))
	}

	return SettingsFormResources{
		settings: settings,
		form: NewForm("Done",
			FormItem{Label: "CPU Limit", Field: cpuField},
			FormItem{Label: "Memory Limit (MB)", Field: memoryField},
		),
	}
}

func (m SettingsFormResources) Title() string {
	return "Resources"
}

func (m SettingsFormResources) Init() tea.Cmd {
	return nil
}

func (m SettingsFormResources) Update(msg tea.Msg) (SettingsSection, tea.Cmd) {
	var (
		action FormAction
		cmd    tea.Cmd
	)
	m.form, action, cmd = m.form.Update(msg)

	switch action {
	case FormSubmitted:
		m.settings.Resources.CPUs, _ = strconv.Atoi(m.form.TextField(resourcesCPUField).Value())
		m.settings.Resources.MemoryMB, _ = strconv.Atoi(m.form.TextField(resourcesMemoryField).Value())
		return m, func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	case FormCancelled:
		return m, func() tea.Msg { return SettingsSectionCancelMsg{} }
	}

	return m, cmd
}

func (m SettingsFormResources) View() string {
	return m.form.View()
}
