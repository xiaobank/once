package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type FormField interface {
	Update(tea.Msg) tea.Cmd
	View() string
	Focus() tea.Cmd
	Blur()
	SetWidth(int)
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

// CheckboxField

type CheckboxField struct {
	label      string
	checked    bool
	disabledFn func() (disabled bool, text string)
}

func NewCheckboxField(label string, checked bool) *CheckboxField {
	return &CheckboxField{label: label, checked: checked}
}

func (f *CheckboxField) Checked() bool {
	return f.checked
}

func (f *CheckboxField) SetDisabledWhen(fn func() (disabled bool, text string)) {
	f.disabledFn = fn
}

func (f *CheckboxField) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, key.NewBinding(key.WithKeys("space"))) {
			if f.disabledFn != nil {
				if disabled, _ := f.disabledFn(); disabled {
					return nil
				}
			}
			f.checked = !f.checked
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
		return "[x] " + f.label
	}
	return "[ ] " + f.label
}

func (f *CheckboxField) Focus() tea.Cmd { return nil }
func (f *CheckboxField) Blur()          {}
func (f *CheckboxField) SetWidth(int)   {}

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
	items       []FormItem
	submitLabel string
	focused     int
	width       int
}

func NewForm(submitLabel string, items ...FormItem) Form {
	f := Form{
		items:       items,
		submitLabel: submitLabel,
	}

	if len(items) > 0 {
		items[0].Field.Focus()
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
		label := Styles.Label.Render(item.Label)
		field := Styles.Focus(Styles.Input, f.focused == i).
			Render(item.Field.View())
		parts = append(parts, label, field)
	}

	submitButton := Styles.Focus(Styles.ButtonPrimary, f.focused == f.submitIndex()).
		Render(f.submitLabel)
	cancelButton := Styles.Focus(Styles.Button, f.focused == f.cancelIndex()).
		Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, submitButton, cancelButton)
	parts = append(parts, "", buttons)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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
	return f.focusCurrent()
}

func (f Form) focusPrev() (Form, tea.Cmd) {
	f.blurCurrent()
	f.focused = (f.focused - 1 + f.totalCount()) % f.totalCount()
	return f.focusCurrent()
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

func (f Form) handleEnter() (Form, FormAction, tea.Cmd) {
	switch {
	case f.focused < len(f.items):
		form, cmd := f.focusNext()
		return form, FormNoAction, cmd
	case f.focused == f.submitIndex():
		return f, FormSubmitted, nil
	case f.focused == f.cancelIndex():
		return f, FormCancelled, nil
	}
	return f, FormNoAction, nil
}

func (f Form) submitIndex() int { return len(f.items) }
func (f Form) cancelIndex() int { return len(f.items) + 1 }
func (f Form) totalCount() int  { return len(f.items) + 2 }
