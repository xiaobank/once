package ui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basecamp/once/internal/docker"
	"github.com/basecamp/once/internal/metrics"
)

func TestFormatDuration(t *testing.T) {
	t.Run("seconds", func(t *testing.T) {
		assert.Equal(t, "0s", formatDuration(0))
		assert.Equal(t, "1s", formatDuration(1*time.Second))
		assert.Equal(t, "45s", formatDuration(45*time.Second))
		assert.Equal(t, "59s", formatDuration(59*time.Second))
	})

	t.Run("minutes", func(t *testing.T) {
		assert.Equal(t, "1m", formatDuration(1*time.Minute))
		assert.Equal(t, "30m", formatDuration(30*time.Minute))
		assert.Equal(t, "59m", formatDuration(59*time.Minute))
		assert.Equal(t, "1m", formatDuration(1*time.Minute+30*time.Second))
	})

	t.Run("hours", func(t *testing.T) {
		assert.Equal(t, "1h", formatDuration(1*time.Hour))
		assert.Equal(t, "2h", formatDuration(2*time.Hour))
		assert.Equal(t, "3h 45m", formatDuration(3*time.Hour+45*time.Minute))
		assert.Equal(t, "23h 59m", formatDuration(23*time.Hour+59*time.Minute))
	})

	t.Run("days", func(t *testing.T) {
		assert.Equal(t, "1d", formatDuration(24*time.Hour))
		assert.Equal(t, "2d", formatDuration(48*time.Hour))
		assert.Equal(t, "1d 1h", formatDuration(25*time.Hour))
		assert.Equal(t, "2d 2h", formatDuration(50*time.Hour))
		assert.Equal(t, "7d 12h", formatDuration(7*24*time.Hour+12*time.Hour))
	})
}

func TestPanelIndexAtY(t *testing.T) {
	d := testDashboard(3)
	slotHeight := PanelHeight + 2 // transition + body + transition
	titleHeight := 1              // title rule + blank line

	t.Run("click on first panel", func(t *testing.T) {
		idx, ok := d.panelIndexAtY(titleHeight)
		require.True(t, ok)
		assert.Equal(t, 0, idx)
	})

	t.Run("click on second panel", func(t *testing.T) {
		idx, ok := d.panelIndexAtY(titleHeight + slotHeight)
		require.True(t, ok)
		assert.Equal(t, 1, idx)
	})

	t.Run("click on third panel", func(t *testing.T) {
		idx, ok := d.panelIndexAtY(titleHeight + slotHeight*2)
		require.True(t, ok)
		assert.Equal(t, 2, idx)
	})

	t.Run("click above panels", func(t *testing.T) {
		_, ok := d.panelIndexAtY(0)
		assert.False(t, ok)
	})

	t.Run("click below all panels", func(t *testing.T) {
		_, ok := d.panelIndexAtY(titleHeight + slotHeight*3)
		assert.False(t, ok)
	})

	t.Run("click with scroll offset", func(t *testing.T) {
		scrolled := testDashboard(3)
		scrolled.width = 80
		scrolled.height = 20
		scrolled.updateViewportSize()
		scrolled.rebuildViewportContent()
		scrolled.viewport.SetYOffset(slotHeight)

		// Y is in screen coordinates; with offset, the first visible panel is index 1
		idx, ok := scrolled.panelIndexAtY(titleHeight)
		require.True(t, ok)
		assert.Equal(t, 1, idx)
	})
}

func TestDashboardMouseClickSelectsPanel(t *testing.T) {
	d := testDashboard(3)
	d.width = 80
	d.height = 40
	d.updateViewportSize()
	d.rebuildViewportContent()

	assert.Equal(t, 0, d.selectedIndex)

	slotHeight := PanelHeight + 2
	titleHeight := 1

	// Click on the second panel
	msg := tea.MouseClickMsg{X: 10, Y: titleHeight + slotHeight, Button: tea.MouseLeft}
	result, _ := d.Update(msg)
	d = result.(Dashboard)
	assert.Equal(t, 1, d.selectedIndex)

	// Click on the third panel
	msg = tea.MouseClickMsg{X: 10, Y: titleHeight + slotHeight*2, Button: tea.MouseLeft}
	result, _ = d.Update(msg)
	d = result.(Dashboard)
	assert.Equal(t, 2, d.selectedIndex)
}

func TestDashboardMouseClickWithScroll(t *testing.T) {
	d := testDashboard(3)
	d.width = 80
	d.height = 20 // short enough that not all panels fit
	d.updateViewportSize()
	d.rebuildViewportContent()

	slotHeight := PanelHeight + 2
	titleHeight := 1

	// Scroll down so the second panel is at the top of the viewport
	d.viewport.SetYOffset(slotHeight)

	// Click at the top of the visible area — should hit panel 1 (index 1)
	msg := tea.MouseClickMsg{X: 10, Y: titleHeight, Button: tea.MouseLeft}
	result, _ := d.Update(msg)
	d = result.(Dashboard)
	assert.Equal(t, 1, d.selectedIndex)
}

// Helpers

func testDashboard(numApps int) Dashboard {
	apps := make([]*docker.Application, numApps)
	for i := range numApps {
		apps[i] = &docker.Application{
			Settings: docker.ApplicationSettings{
				Name: fmt.Sprintf("app-%d", i),
				Host: fmt.Sprintf("app-%d.example.com", i),
			},
		}
	}

	scraper := metrics.NewMetricsScraper(metrics.ScraperSettings{})
	dockerScraper := &docker.Scraper{}

	return NewDashboard(nil, apps, 0, scraper, dockerScraper)
}
