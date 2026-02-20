package main

import (
	"time"

	"github.com/basecamp/gliff/tui"
)

type HelloModel struct {
	currentTime time.Time
}

type tickMsg time.Time

func (m *HelloModel) Init() tui.Cmd {
	return m.scheduleTick()
}

func (m *HelloModel) Update(msg tui.Msg) tui.Cmd {
	switch msg := msg.(type) {
	case tui.KeyMsg:
		if msg.Type == tui.KeyCtrlC {
			return tui.Quit
		}
	case tickMsg:
		m.currentTime = time.Time(msg)
		return m.scheduleTick()
	}
	return nil
}

func (m *HelloModel) Render() string {
	if m.currentTime.IsZero() {
		return ""
	}

	return "Hello! It's " + m.currentTime.Format("15:04:05")
}

func (m *HelloModel) scheduleTick() tui.Cmd {
	return tui.Every(time.Second, func() tui.Msg {
		return tickMsg(time.Now())
	})
}

func main() {
	m := &HelloModel{}
	app := tui.NewApplication(m)

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
