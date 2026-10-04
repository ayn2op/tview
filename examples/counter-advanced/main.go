package main

import (
	"log"
	"strconv"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/row"
	"github.com/ayn2op/tview/text"
)

type (
	incrementMsg struct{}
	decrementMsg struct{}
	resetMsg     struct{}
)

type focus int

const (
	focusIncrement focus = iota
	focusDecrement
	focusReset
)

type model struct {
	count int
	focus focus
}

var _ tview.Model[model] = model{}

func (m model) Init() tview.Cmd {
	return tview.SetTitle("Counter")
}

func (m model) Update(msg tview.Msg) (model, tview.Cmd) {
	switch msg := msg.(type) {
	case resetMsg:
		m.count = 0
		m.focus = focusReset
	case decrementMsg:
		m.count--
		m.focus = focusDecrement
	case incrementMsg:
		m.count++
		m.focus = focusIncrement
	case tview.KeyMsg:
		if keybind.String(msg) == "q" {
			return m, tview.Quit()
		}
	}
	return m, nil
}

func (m model) View() tview.Widget {
	return column.New(
		text.New("Count: "+strconv.Itoa(m.count)),
		row.New(
			button.New().
				Label("Reset").
				Focused(m.focus == focusReset).
				OnClick(resetMsg{}),
			button.New().
				Label("Decrement").
				Focused(m.focus == focusDecrement).
				OnClick(decrementMsg{}),
			button.New().
				Label("Increment").
				Focused(m.focus == focusIncrement).
				OnClick(incrementMsg{}),
		),
	)
}

func main() {
	if err := tview.NewApplication(model{}).Run(); err != nil {
		log.Fatal(err)
	}
}
