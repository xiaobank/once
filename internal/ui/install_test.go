package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/basecamp/once/internal/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_SubmitTriggersActivity(t *testing.T) {
	m := newTestInstall()
	assert.Equal(t, installStateForm, m.state)

	m, _ = updateInstall(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = updateInstall(m, InstallFormSubmitMsg{ImageRef: "nginx:latest", Hostname: "app.example.com"})
	assert.Equal(t, installStateActivity, m.state)
}

func TestInstall_SuccessNavigatesToApp(t *testing.T) {
	m := newTestInstall()
	app := &docker.Application{}

	_, cmd := updateInstall(m, InstallActivityDoneMsg{App: app})
	require.NotNil(t, cmd)

	msg := cmd()
	navMsg, ok := msg.(NavigateToAppMsg)
	require.True(t, ok, "expected NavigateToAppMsg, got %T", msg)
	assert.Equal(t, app, navMsg.App)
}

func TestInstall_FailureReturnsToFormWithError(t *testing.T) {
	m := newTestInstall()

	// Fill the form fields before submitting
	fillInstallForm(&m.form, "nginx:latest", "app.example.com")

	// Submit to enter activity state
	m, _ = updateInstall(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = updateInstall(m, InstallFormSubmitMsg{ImageRef: "nginx:latest", Hostname: "app.example.com"})
	assert.Equal(t, installStateActivity, m.state)

	// Simulate failure
	installErr := errors.New("connection refused")
	m, cmd := updateInstall(m, InstallActivityFailedMsg{Err: installErr})

	assert.NotNil(t, cmd, "expected logo Init cmd on failure return")
	assert.Equal(t, installStateForm, m.state)
	assert.Equal(t, installErr, m.err)
	assert.Contains(t, m.View(), "Error: connection refused")

	// Form field values are preserved
	assert.Equal(t, "nginx:latest", m.form.ImageRef())
	assert.Equal(t, "app.example.com", m.form.Hostname())
}

func TestInstall_ErrorClearsOnKeypress(t *testing.T) {
	m := newTestInstall()
	m.err = errors.New("some error")

	m, _ = updateInstall(m, keyPressMsg("a"))
	assert.Nil(t, m.err)
}

func TestInstall_EscNavigatesToDashboard(t *testing.T) {
	m := newTestInstall()

	_, cmd := updateInstall(m, keyPressMsg("esc"))
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(NavigateToDashboardMsg)
	assert.True(t, ok, "expected NavigateToDashboardMsg, got %T", msg)
}

func TestInstall_CancelNavigatesToDashboard(t *testing.T) {
	m := newTestInstall()

	_, cmd := updateInstall(m, InstallFormCancelMsg{})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(NavigateToDashboardMsg)
	assert.True(t, ok, "expected NavigateToDashboardMsg, got %T", msg)
}

func TestInstall_EscQuitsWhenImageRefSet(t *testing.T) {
	m := NewInstall(nil, "nginx:latest")

	_, cmd := updateInstall(m, keyPressMsg("esc"))
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(QuitMsg)
	assert.True(t, ok, "expected QuitMsg, got %T", msg)
}

func TestInstall_EscNavigatesToDashboardEvenWithFieldsFilled(t *testing.T) {
	m := newTestInstall()
	fillInstallForm(&m.form, "nginx:latest", "app.example.com")

	_, cmd := updateInstall(m, keyPressMsg("esc"))
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(NavigateToDashboardMsg)
	assert.True(t, ok, "expected NavigateToDashboardMsg, got %T", msg)
}

func TestInstall_CancelQuitsWhenImageRefSet(t *testing.T) {
	m := NewInstall(nil, "nginx:latest")

	_, cmd := updateInstall(m, InstallFormCancelMsg{})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(QuitMsg)
	assert.True(t, ok, "expected QuitMsg, got %T", msg)
}

func TestInstall_ShowsLogoAndHidesTitleWhenNoApps(t *testing.T) {
	m := NewInstall(nil, "")
	m, _ = updateInstall(m, tea.WindowSizeMsg{Width: 80, Height: 40})

	view := m.View()
	assert.Contains(t, view, "██████╗")
	assert.NotContains(t, view, "ONCE · install")
}

func TestInstall_ShowsTitleAndHidesLogoWhenAppsExist(t *testing.T) {
	ns := newTestNamespace("myapp")
	m := NewInstall(ns, "")
	m, _ = updateInstall(m, tea.WindowSizeMsg{Width: 80, Height: 40})

	view := m.View()
	assert.NotContains(t, view, "██████╗")
	assert.Contains(t, view, "ONCE · install")
}

func TestInstall_FailureRestartsLogoOnlyWhenNoApps(t *testing.T) {
	noApps := NewInstall(nil, "")
	noApps, _ = updateInstall(noApps, tea.WindowSizeMsg{Width: 80, Height: 40})
	noApps, _ = updateInstall(noApps, InstallFormSubmitMsg{ImageRef: "nginx:latest", Hostname: "app.example.com"})
	_, cmd := updateInstall(noApps, InstallActivityFailedMsg{Err: errors.New("fail")})
	assert.NotNil(t, cmd)

	withApps := NewInstall(newTestNamespace("myapp"), "")
	withApps, _ = updateInstall(withApps, tea.WindowSizeMsg{Width: 80, Height: 40})
	withApps, _ = updateInstall(withApps, InstallFormSubmitMsg{ImageRef: "nginx:latest", Hostname: "app.example.com"})
	_, cmd = updateInstall(withApps, InstallActivityFailedMsg{Err: errors.New("fail")})
	assert.Nil(t, cmd)
}

// Helpers

func newTestInstall() Install {
	return NewInstall(nil, "")
}

func newTestNamespace(appNames ...string) *docker.Namespace {
	ns, err := docker.NewNamespace("test")
	if err != nil {
		panic(err)
	}
	for _, name := range appNames {
		ns.AddApplication(docker.ApplicationSettings{Name: name})
	}
	return ns
}

func fillInstallForm(form *InstallForm, imageRef, hostname string) {
	installTypeText(form, imageRef)
	installPressEnter(form)
	installTypeText(form, hostname)
}

func updateInstall(m Install, msg tea.Msg) (Install, tea.Cmd) {
	comp, cmd := m.Update(msg)
	return comp.(Install), cmd
}
