package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

type Help struct {
	model  help.Model
	prefix string
}

func NewHelp() Help {
	return Help{
		model:  help.New(),
		prefix: zone.NewPrefix(),
	}
}

func (h *Help) SetWidth(w int) {
	h.model.SetWidth(w)
}

func (h Help) View(k help.KeyMap) string {
	bindings := k.ShortHelp()
	if len(bindings) == 0 {
		return ""
	}

	var b strings.Builder
	var totalWidth int
	separator := h.model.Styles.ShortSeparator.Inline(true).Render(h.model.ShortSeparator)

	for i, kb := range bindings {
		if !kb.Enabled() {
			continue
		}

		var sep string
		if totalWidth > 0 {
			sep = separator
		}

		item := h.model.Styles.ShortKey.Inline(true).Render(kb.Help().Key) + " " +
			h.model.Styles.ShortDesc.Inline(true).Render(kb.Help().Desc)
		str := sep + zone.Mark(h.zoneID(i), item)
		w := lipgloss.Width(str)

		if tail, ok := h.shouldAddItem(totalWidth, w); !ok {
			if tail != "" {
				b.WriteString(tail)
			}
			break
		}

		totalWidth += w
		b.WriteString(str)
	}

	return b.String()
}

func (h Help) Update(msg tea.MouseClickMsg, k help.KeyMap) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return nil
	}

	for i, kb := range k.ShortHelp() {
		if !kb.Enabled() {
			continue
		}
		if zi := zone.Get(h.zoneID(i)); zi != nil && zi.InBounds(msg) {
			return func() tea.Msg { return keyPressFromBinding(kb) }
		}
	}

	return nil
}

// Private

func (h Help) zoneID(i int) string {
	return fmt.Sprintf("%shelp_%d", h.prefix, i)
}

func (h Help) shouldAddItem(totalWidth, width int) (tail string, ok bool) {
	w := h.model.Width()
	if w > 0 && totalWidth+width > w {
		tail = " " + h.model.Styles.Ellipsis.Inline(true).Render(h.model.Ellipsis)
		if totalWidth+lipgloss.Width(tail) < w {
			return tail, false
		}
	}
	return "", true
}

// Helpers

func keyPressFromBinding(b key.Binding) tea.KeyPressMsg {
	keys := b.Keys()
	if len(keys) == 0 {
		return tea.KeyPressMsg{}
	}

	k := keys[0]

	switch k {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	}

	if len(k) == 1 {
		r := rune(k[0])
		return tea.KeyPressMsg{Code: r, Text: k}
	}

	return tea.KeyPressMsg{}
}
