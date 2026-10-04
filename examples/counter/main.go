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

type incrementMsg struct{}

type model struct{ count int }

var _ tview.Model[model] = model{}

func (m model) Init() tview.Cmd {
	return tview.SetTitle("Counter")
}

func (m model) Update(msg tview.Msg) (model, tview.Cmd) {
	switch msg := msg.(type) {
	case incrementMsg:
		m.count++
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
				Label("Increment").
				OnClick(incrementMsg{}).
				Focused(true),
		),
	)
}

func main() {
	if err := tview.NewApplication(model{}).Run(); err != nil {
		log.Fatal(err)
	}
}
