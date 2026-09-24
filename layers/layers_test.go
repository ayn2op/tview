package layers

import (
	"slices"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/flex"
	"github.com/ayn2op/tview/grid"
	"github.com/ayn2op/tview/tabs"
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

func TestLayersInit(t *testing.T) {
	for name, build := range map[string]func(active, inactive *initModel) *Layers{
		"active layer": func(active, inactive *initModel) *Layers {
			return New().AddLayer(active, WithName("active")).
				AddLayer(inactive, WithName("hidden"), WithVisible(false)).
				AddLayer(inactive, WithName("disabled"), WithEnabled(false))
		},
		"nested containers": func(active, inactive *initModel) *Layers {
			child := tabs.NewModel([]tabs.Tab{active, inactive})
			layout := grid.NewModel().AddItem(child, 0, 0, 1, 1, 0, 0, true)
			return New().AddLayer(flex.NewModel().AddItem(layout, 0, 1, true))
		},
	} {
		t.Run(name, func(t *testing.T) {
			active := &initModel{Box: tview.NewBox()}
			inactive := &initModel{Box: tview.NewBox()}
			cmd := build(active, inactive).Init()
			if active.calls != 1 || inactive.calls != 0 {
				t.Fatalf("Init calls: active=%d, inactive=%d", active.calls, inactive.calls)
			}
			if cmd == nil || cmd() != active {
				t.Fatal("active child's startup command was not returned")
			}
		})
	}
	t.Run("empty", func(t *testing.T) {
		if New().Init() != nil {
			t.Fatal("empty layers returned a startup command")
		}
	})
}

func TestLayerOrder(t *testing.T) {
	layers := New()
	for _, name := range []string{"a", "b", "c"} {
		layers.AddLayer(tview.NewBox(), WithName(name))
	}

	assertNames := func(want ...string) {
		t.Helper()
		if got := layers.GetLayerNames(false); !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	assertNames("c", "b", "a")
	layers.SendToFront("a")
	assertNames("a", "c", "b")
	layers.SendToBack("a")
	assertNames("c", "b", "a")
	layers.RemoveLayer("b")
	assertNames("c", "a")
	layers.AddLayer(tview.NewBox(), WithName("a"))
	assertNames("a", "c")
}
