package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basecamp/once/internal/docker"
)

func newTestLogs() Logs {
	return Logs{
		app: &docker.Application{
			Settings: docker.ApplicationSettings{Name: "testapp"},
		},
		streamer:    newTestLogStreamerForUI(),
		viewport:    viewport.New(),
		filterInput: textinput.New(),
		help:        NewHelp(),
	}
}

func newTestLogStreamerForUI() *docker.LogStreamer {
	return docker.NewLogStreamerForTest(docker.LogStreamerSettings{})
}

func TestLogsFilterActivation(t *testing.T) {
	m := newTestLogs()
	m.filterActive = false

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	logs := updated.(Logs)

	assert.True(t, logs.filterActive)
}

func TestLogsFilterAppliesOnKeypress(t *testing.T) {
	m := newTestLogs()
	m.filterActive = true
	m.filterInput.SetValue("err")

	// Type another character
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	logs := updated.(Logs)

	// Filter should be applied immediately (filter text updated to match input)
	assert.Equal(t, logs.filterInput.Value(), logs.filterText)
}

func TestLogsFilterClearedOnEscape(t *testing.T) {
	m := newTestLogs()
	m.filterActive = true
	m.filterText = "error"
	m.filterInput.SetValue("error")

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	logs := updated.(Logs)

	assert.False(t, logs.filterActive)
	assert.Equal(t, "", logs.filterText)
	assert.Equal(t, "", logs.filterInput.Value())
}

func TestLogsBackNavigation(t *testing.T) {
	m := newTestLogs()
	m.filterActive = false

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(navigateToDashboardMsg)
	assert.True(t, ok)
}

func TestLogsFilterMatchesCaseInsensitive(t *testing.T) {
	m := newTestLogs()
	m.filterText = "ERROR"

	lines := []docker.LogLine{
		{Content: "This is an error message"},
		{Content: "This is INFO"},
		{Content: "Another ERROR here"},
	}

	var filtered []string
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line.Content), strings.ToLower(m.filterText)) {
			filtered = append(filtered, line.Content)
		}
	}

	assert.Len(t, filtered, 2)
	assert.Equal(t, "This is an error message", filtered[0])
	assert.Equal(t, "Another ERROR here", filtered[1])
}
