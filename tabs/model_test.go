package tabs

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type initModel struct {
	*tview.Box
	calls int
}

func (m *initModel) Init() tview.Cmd {
	m.calls++
	return func() tview.Msg { return m }
}

func (*initModel) Label() string { return "tab" }

func TestModelInit(t *testing.T) {
	t.Run("active tab", func(t *testing.T) {
		active := &initModel{Box: tview.NewBox()}
		inactive := &initModel{Box: tview.NewBox()}
		cmd := NewModel([]Tab{active, inactive}).Init()
		if active.calls != 1 || inactive.calls != 0 {
			t.Fatalf("Init calls: active=%d, inactive=%d", active.calls, inactive.calls)
		}
		if cmd == nil || cmd() != active {
			t.Fatal("active child's startup command was not returned")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if NewModel(nil).Init() != nil {
			t.Fatal("empty model returned a startup command")
		}
	})
}

func TestModelUpdate(t *testing.T) {
	t.Run("tab activation", func(t *testing.T) {
		first := &initModel{Box: tview.NewBox()}
		second := &initModel{Box: tview.NewBox()}
		model := NewModel([]Tab{first, second})
		model.Init()
		cmd := model.Update(tcell.NewEventKey(tcell.KeyCtrlL, "", tcell.ModNone))
		if first.calls != 1 || second.calls != 1 || cmd == nil || cmd() != second {
			t.Fatal("switching tabs did not initialize the newly active tab")
		}
		model.Update(tcell.NewEventKey(tcell.KeyCtrlL, "", tcell.ModNone))
		if second.calls != 1 {
			t.Fatal("navigation at the last tab reinitialized it")
		}
	})
}
