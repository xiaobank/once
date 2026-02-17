package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

type FormField interface {
	Update(tea.Msg) tea.Cmd
	View() string
	Focus() tea.Cmd
	Blur()
	SetWidth(int)
	IsFocusable() bool
}

// TextField

type TextField struct {
	input      textinput.Model
	digitsOnly bool
}

func NewTextField(placeholder string) *TextField {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Prompt = ""
	input.CharLimit = 256
	return &TextField{input: input}
}

func (f *TextField) Value() string {
	return f.input.Value()
}

func (f *TextField) SetValue(v string) {
	f.input.SetValue(v)
}

func (f *TextField) SetPlaceholder(p string) {
	f.input.Placeholder = p
}

func (f *TextField) SetCharLimit(n int) {
	f.input.CharLimit = n
}

func (f *TextField) SetDigitsOnly(v bool) {
	f.digitsOnly = v
}

func (f *TextField) SetEchoPassword() {
	f.input.EchoMode = textinput.EchoPassword
}

func (f *TextField) Update(msg tea.Msg) tea.Cmd {
	if f.digitsOnly {
		if msg, ok := msg.(tea.KeyMsg); ok {
			if text := msg.Key().Text; text != "" && (text[0] < '0' || text[0] > '9') {
				return nil
			}
		}
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd
}

func (f *TextField) View() string {
	return f.input.View()
}

func (f *TextField) Focus() tea.Cmd {
	return f.input.Focus()
}

func (f *TextField) Blur() {
	f.input.Blur()
}

func (f *TextField) SetWidth(w int) {
	f.input.SetWidth(w)
}
func (f *TextField) IsFocusable() bool { return true }

// CheckboxField

type CheckboxField struct {
	label      string
	checked    bool
	disabledFn func() (disabled bool, text string)
}

func NewCheckboxField(label string, checked bool) *CheckboxField {
	return &CheckboxField{label: label, checked: checked}
}

// StaticField

type StaticField struct {
	value   string
	styleFn func(string) string
}

func NewStaticField(value string, styleFn func(string) string) *StaticField {
	return &StaticField{value: value, styleFn: styleFn}
}

func (f *StaticField) Value() string {
	return f.value
}

func (f *StaticField) SetValue(v string) {
	f.value = v
}

func (f *StaticField) Update(tea.Msg) tea.Cmd { return nil }
func (f *StaticField) View() string           { return f.styleFn(f.value) }
func (f *StaticField) Focus() tea.Cmd         { return nil }
func (f *StaticField) Blur()                  {}
func (f *StaticField) SetWidth(int)           {}
func (f *StaticField) IsFocusable() bool      { return false }

func (f *CheckboxField) Checked() bool {
	return f.checked
}

func (f *CheckboxField) SetDisabledWhen(fn func() (disabled bool, text string)) {
	f.disabledFn = fn
}

func (f *CheckboxField) Toggle() {
	if f.disabledFn != nil {
		if disabled, _ := f.disabledFn(); disabled {
			return
		}
	}
	f.checked = !f.checked
}

func (f *CheckboxField) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, key.NewBinding(key.WithKeys("space"))) {
			f.Toggle()
		}
	}
	return nil
}

func (f *CheckboxField) View() string {
	if f.disabledFn != nil {
		if disabled, text := f.disabledFn(); disabled {
			return text
		}
	}

	if f.checked {
		return "[✓] " + f.label
	}
	return "[ ] " + f.label
}

func (f *CheckboxField) Focus() tea.Cmd    { return nil }
func (f *CheckboxField) Blur()             {}
func (f *CheckboxField) SetWidth(int)      {}
func (f *CheckboxField) IsFocusable() bool { return true }

// FormActionButton

type FormActionButton struct {
	Label   string
	OnPress func() tea.Msg
}

// Form

type FormAction int

const (
	FormNoAction FormAction = iota
	FormSubmitted
	FormCancelled
)

type FormItem struct {
	Label string
	Field FormField
}

type Form struct {
	items        []FormItem
	submitLabel  string
	actionButton *FormActionButton
	focused      int
	width        int
	prefix       string
}

func NewForm(submitLabel string, items ...FormItem) Form {
	f := Form{
		items:       items,
		submitLabel: submitLabel,
		prefix:      zone.NewPrefix(),
	}

	// Find first focusable field and focus it
	for i, item := range items {
		if item.Field.IsFocusable() {
			f.focused = i
			item.Field.Focus()
			break
		}
	}

	return f
}

func (f Form) Init() tea.Cmd {
	return nil
}

func (f Form) Update(msg tea.Msg) (Form, FormAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		inputWidth := min(f.width-4, 60)
		for _, item := range f.items {
			item.Field.SetWidth(inputWidth)
		}

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return f.handleMouseClick(msg)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			form, cmd := f.focusNext()
			return form, FormNoAction, cmd
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			form, cmd := f.focusPrev()
			return form, FormNoAction, cmd
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return f.handleEnter()
		}
	}

	if f.focused < len(f.items) {
		cmd := f.items[f.focused].Field.Update(msg)
		return f, FormNoAction, cmd
	}

	return f, FormNoAction, nil
}

func (f Form) View() string {
	var parts []string

	for i, item := range f.items {
		// Static fields don't have a label or border
		if _, isStatic := item.Field.(*StaticField); isStatic {
			parts = append(parts, item.Field.View())
			continue
		}
		label := Styles.Label.Render(item.Label)
		field := Styles.Focus(Styles.Input, f.focused == i).
			Render(item.Field.View())
		parts = append(parts, label, zone.Mark(f.fieldZoneID(i), field))
	}

	submitButton := zone.Mark(f.submitZoneID(),
		Styles.Focus(Styles.ButtonPrimary, f.focused == f.submitIndex()).
			Render(f.submitLabel))

	buttonParts := []string{submitButton}

	if f.actionButton != nil {
		actionBtn := zone.Mark(f.actionZoneID(),
			Styles.Focus(Styles.Button, f.focused == f.actionIndex()).
				Render(f.actionButton.Label))
		buttonParts = append(buttonParts, actionBtn)
	}

	cancelButton := zone.Mark(f.cancelZoneID(),
		Styles.Focus(Styles.Button, f.focused == f.cancelIndex()).
			Render("Cancel"))

	buttonParts = append(buttonParts, cancelButton)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, buttonParts...)
	parts = append(parts, "", buttons)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (f *Form) SetActionButton(label string, onPress func() tea.Msg) {
	f.actionButton = &FormActionButton{Label: label, OnPress: onPress}
}

func (f Form) Focused() int {
	return f.focused
}

func (f Form) Field(i int) FormField {
	return f.items[i].Field
}

func (f Form) TextField(i int) *TextField {
	return f.items[i].Field.(*TextField)
}

func (f Form) CheckboxField(i int) *CheckboxField {
	return f.items[i].Field.(*CheckboxField)
}

// Private

func (f Form) focusNext() (Form, tea.Cmd) {
	f.blurCurrent()
	f.focused = (f.focused + 1) % f.totalCount()
	return f.focusToNextFocusable()
}

func (f Form) focusPrev() (Form, tea.Cmd) {
	f.blurCurrent()
	f.focused = (f.focused - 1 + f.totalCount()) % f.totalCount()
	return f.focusToNextFocusable()
}

func (f *Form) blurCurrent() {
	if f.focused < len(f.items) {
		f.items[f.focused].Field.Blur()
	}
}

func (f Form) focusCurrent() (Form, tea.Cmd) {
	if f.focused < len(f.items) {
		cmd := f.items[f.focused].Field.Focus()
		return f, cmd
	}
	return f, nil
}

func (f Form) focusToNextFocusable() (Form, tea.Cmd) {
	// Skip non-focusable fields and buttons (which are always focusable)
	start := f.focused
	for {
		if f.focused < len(f.items) {
			if f.items[f.focused].Field.IsFocusable() {
				return f.focusCurrent()
			}
			// Skip to next
			f.focused = (f.focused + 1) % f.totalCount()
		} else {
			// We've reached buttons, which are always focusable
			return f, nil
		}

		// Prevent infinite loop
		if f.focused == start {
			return f, nil
		}
	}
}

func (f Form) handleEnter() (Form, FormAction, tea.Cmd) {
	switch {
	case f.focused < len(f.items):
		form, cmd := f.focusNext()
		return form, FormNoAction, cmd
	case f.actionButton != nil && f.focused == f.actionIndex():
		return f, FormNoAction, func() tea.Msg { return f.actionButton.OnPress() }
	case f.focused == f.submitIndex():
		return f, FormSubmitted, nil
	case f.focused == f.cancelIndex():
		return f, FormCancelled, nil
	}
	return f, FormNoAction, nil
}

func (f Form) handleMouseClick(msg tea.MouseClickMsg) (Form, FormAction, tea.Cmd) {
	for i := range f.items {
		if zi := zone.Get(f.fieldZoneID(i)); zi != nil && zi.InBounds(msg) {
			// Only focus if the field is focusable
			if !f.items[i].Field.IsFocusable() {
				// For non-focusable fields, just handle clicks (e.g., checkbox toggle) but don't change focus
				if cb, ok := f.items[i].Field.(*CheckboxField); ok {
					cb.Toggle()
				}
				return f, FormNoAction, nil
			}
			form, focusCmd := f.focusIndex(i)
			if cb, ok := f.items[i].Field.(*CheckboxField); ok {
				cb.Toggle()
			}
			return form, FormNoAction, focusCmd
		}
	}

	if f.actionButton != nil {
		if zi := zone.Get(f.actionZoneID()); zi != nil && zi.InBounds(msg) {
			return f, FormNoAction, func() tea.Msg { return f.actionButton.OnPress() }
		}
	}

	if zi := zone.Get(f.submitZoneID()); zi != nil && zi.InBounds(msg) {
		return f, FormSubmitted, nil
	}

	if zi := zone.Get(f.cancelZoneID()); zi != nil && zi.InBounds(msg) {
		return f, FormCancelled, nil
	}

	return f, FormNoAction, nil
}

func (f Form) focusIndex(i int) (Form, tea.Cmd) {
	f.blurCurrent()
	f.focused = i
	return f.focusCurrent()
}

func (f Form) zoneID(name string) string {
	return fmt.Sprintf("%s%s", f.prefix, name)
}

func (f Form) fieldZoneID(i int) string {
	return f.zoneID(fmt.Sprintf("field_%d", i))
}

func (f Form) actionZoneID() string { return f.zoneID("action") }
func (f Form) submitZoneID() string { return f.zoneID("submit") }
func (f Form) cancelZoneID() string { return f.zoneID("cancel") }

func (f Form) submitIndex() int { return len(f.items) }

func (f Form) actionIndex() int { return len(f.items) + 1 }

func (f Form) cancelIndex() int {
	if f.actionButton != nil {
		return len(f.items) + 2
	}
	return len(f.items) + 1
}

func (f Form) totalCount() int {
	return f.cancelIndex() + 1
}
