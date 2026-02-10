package ui

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testKeyMap struct {
	bindings []key.Binding
}

func (k testKeyMap) ShortHelp() []key.Binding { return k.bindings }
func (k testKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.bindings} }

func TestHelp_ViewMatchesHelpModel(t *testing.T) {
	km := testKeyMap{bindings: []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "logs")),
	}}

	h := NewHelp()
	standard := help.New()

	got := zone.Scan(h.View(km))
	want := standard.ShortHelpView(km.ShortHelp())

	assert.Equal(t, want, got)
}

func TestHelp_ViewWithDisabledBinding(t *testing.T) {
	disabled := key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "hidden"))
	disabled.SetEnabled(false)

	km := testKeyMap{bindings: []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
		disabled,
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "logs")),
	}}

	h := NewHelp()
	standard := help.New()

	got := zone.Scan(h.View(km))
	want := standard.ShortHelpView(km.ShortHelp())

	assert.Equal(t, want, got)
}

func TestHelp_ClickGeneratesKeyPress(t *testing.T) {
	km := testKeyMap{bindings: []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "logs")),
	}}

	h := NewHelp()
	zone.Scan(h.View(km))
	time.Sleep(15 * time.Millisecond)

	zi := zone.Get(h.zoneID(1))
	require.False(t, zi.IsZero())

	cmd := h.Update(tea.MouseClickMsg{X: zi.StartX, Y: zi.StartY, Button: tea.MouseLeft}, km)
	require.NotNil(t, cmd)

	msg := cmd()
	kp, ok := msg.(tea.KeyPressMsg)
	require.True(t, ok)
	assert.Equal(t, 'g', kp.Code)
	assert.Equal(t, "g", kp.Text)
}

func TestHelp_ClickOutOfBounds(t *testing.T) {
	km := testKeyMap{bindings: []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
	}}

	h := NewHelp()
	zone.Scan(h.View(km))
	time.Sleep(15 * time.Millisecond)

	cmd := h.Update(tea.MouseClickMsg{X: 999, Y: 999, Button: tea.MouseLeft}, km)
	assert.Nil(t, cmd)
}

func TestHelp_RightClickIgnored(t *testing.T) {
	km := testKeyMap{bindings: []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
	}}

	h := NewHelp()
	zone.Scan(h.View(km))
	time.Sleep(15 * time.Millisecond)

	zi := zone.Get(h.zoneID(0))
	require.False(t, zi.IsZero())

	cmd := h.Update(tea.MouseClickMsg{X: zi.StartX, Y: zi.StartY, Button: tea.MouseRight}, km)
	assert.Nil(t, cmd)
}

func TestKeyPressFromBinding(t *testing.T) {
	charBinding := key.NewBinding(key.WithKeys("s"))
	msg := keyPressFromBinding(charBinding)
	assert.Equal(t, 's', msg.Code)
	assert.Equal(t, "s", msg.Text)

	escBinding := key.NewBinding(key.WithKeys("esc"))
	msg = keyPressFromBinding(escBinding)
	assert.Equal(t, tea.KeyEscape, msg.Code)
	assert.Empty(t, msg.Text)

	enterBinding := key.NewBinding(key.WithKeys("enter"))
	msg = keyPressFromBinding(enterBinding)
	assert.Equal(t, tea.KeyEnter, msg.Code)

	leftBinding := key.NewBinding(key.WithKeys("left"))
	msg = keyPressFromBinding(leftBinding)
	assert.Equal(t, tea.KeyLeft, msg.Code)

	emptyBinding := key.NewBinding()
	msg = keyPressFromBinding(emptyBinding)
	assert.Equal(t, tea.KeyPressMsg{}, msg)
}
