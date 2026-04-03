package main

import (
	"fmt"

	"github.com/ayn2op/tview"
)

type panicModel struct {
	*tview.Box
}

func (m *panicModel) Update(event tview.Event) tview.Cmd {
	if _, ok := event.(*tview.InitEvent); ok {
		return func() tview.Event {
			panic("boom from anonymous cmd")
		}
	}
	return nil
}

func main() {
	app := tview.NewApplication()
	app.SetRoot(&panicModel{Box: tview.NewBox()})

	if err := app.Run(); err != nil {
		fmt.Printf("Run returned error: %v\n", err)
	}
}
