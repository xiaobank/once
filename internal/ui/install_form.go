package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/docker"
)

const (
	installImageRefField = iota
	installHostnameField
)

type InstallFormSubmitMsg struct {
	ImageRef string
	Hostname string
}

type InstallFormCancelMsg struct{}

type InstallForm struct {
	form        Form
	lastAppName string
	imageRef    string
}

func NewInstallForm(imageRef string) InstallForm {
	var formItems []FormItem

	if imageRef != "" {
		// Expand alias and show static field
		if expanded, ok := imageAliases[imageRef]; ok {
			imageRef = expanded
		}

		styleFunc := func(s string) string {
			return lipgloss.NewStyle().
				Foreground(Colors.Muted).
				Padding(1, 0).
				Width(60).
				Align(lipgloss.Center).
				Render("Installing " + s)
		}

		formItems = append(formItems, FormItem{
			Label: "",
			Field: NewStaticField(imageRef, styleFunc),
		})
	} else {
		formItems = append(formItems, FormItem{
			Label: "Image",
			Field: NewTextField("user/repo:tag"),
		})
	}

	hostnameField := NewTextField("app.example.com")
	formItems = append(formItems, FormItem{
		Label: "Hostname",
		Field: hostnameField,
	})

	form := InstallForm{
		form:     NewForm("Install", formItems...),
		imageRef: imageRef,
	}

	// If imageRef is pre-populated, set the hostname placeholder immediately
	if imageRef != "" {
		form.updateHostnamePlaceholder()
	}

	return form
}

func (m InstallForm) Init() tea.Cmd {
	return nil
}

func (m InstallForm) Update(msg tea.Msg) (InstallForm, tea.Cmd) {
	prev := m.form.Focused()

	var (
		action FormAction
		cmd    tea.Cmd
	)
	m.form, action, cmd = m.form.Update(msg)

	switch action {
	case FormSubmitted:
		return m, func() tea.Msg {
			return InstallFormSubmitMsg{
				ImageRef: m.ImageRef(),
				Hostname: m.form.TextField(installHostnameField).Value(),
			}
		}
	case FormCancelled:
		return m, func() tea.Msg { return InstallFormCancelMsg{} }
	}

	if prev == 0 && m.form.Focused() != 0 && m.imageRef == "" {
		m.expandImageAlias()
		m.updateHostnamePlaceholder()
	}

	return m, cmd
}

func (m InstallForm) View() string {
	return m.form.View()
}

func (m InstallForm) ImageRef() string {
	if m.imageRef != "" {
		return m.imageRef
	}
	return m.form.TextField(installImageRefField).Value()
}

func (m InstallForm) Hostname() string {
	return m.form.TextField(installHostnameField).Value()
}

// Private

var imageAliases = map[string]string{
	"campfire": "ghcr.io/basecamp/once-campfire",
	"fizzy":    "ghcr.io/basecamp/fizzy:main",
}

func (m *InstallForm) expandImageAlias() {
	field := m.form.TextField(installImageRefField)
	if expanded, ok := imageAliases[field.Value()]; ok {
		field.SetValue(expanded)
	}
}

func (m *InstallForm) updateHostnamePlaceholder() {
	appName := docker.NameFromImageRef(m.ImageRef())
	if appName != m.lastAppName && appName != "" {
		m.form.TextField(installHostnameField).SetPlaceholder(appName + ".example.com")
		m.lastAppName = appName
	}
}
