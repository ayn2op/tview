package flex

import (
	"testing"

	"github.com/ayn2op/tview"
)

type initModel struct {
	*tview.Box
	calls int
}

func (m *initModel) Init() tview.Cmd {
	m.calls++
	return func() tview.Msg { return m }
}

func TestModelInit(t *testing.T) {
	t.Run("focused child", func(t *testing.T) {
		active := &initModel{Box: tview.NewBox()}
		inactive := &initModel{Box: tview.NewBox()}
		m := NewModel().AddItem(inactive, 0, 1, false).AddItem(active, 0, 1, true)
		cmd := m.Init()
		if active.calls != 1 || inactive.calls != 0 {
			t.Fatalf("Init calls: active=%d, inactive=%d", active.calls, inactive.calls)
		}
		if cmd == nil || cmd() != active {
			t.Fatal("active child's startup command was not returned")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if NewModel().Init() != nil {
			t.Fatal("empty model returned a startup command")
		}
	})
	t.Run("nil child", func(t *testing.T) {
		if NewModel().AddItem(nil, 0, 1, true).Init() != nil {
			t.Fatal("nil child returned a startup command")
		}
	})
}

func TestLocalFocus(t *testing.T) {
	first := tview.NewBox()
	second := tview.NewBox()
	m := NewModel().
		AddItem(first, 0, 1, true).
		AddItem(second, 0, 1, false)

	if m.Focused() != 0 {
		t.Fatal("first item was not selected")
	}

	m.SetFocus(1)
	if m.Focused() != 1 {
		t.Fatal("second item was not selected")
	}
}
