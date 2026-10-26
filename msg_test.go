package tview

import "testing"

func TestCompactCmds(t *testing.T) {
	first := func() Msg { return 1 }
	second := func() Msg { return 2 }
	cmds := []Cmd{nil, first, nil, second, nil}

	t.Run("keeps the commands in order", func(t *testing.T) {
		if got := compactCmds(cmds...); len(got) != 2 || got[0]() != 1 || got[1]() != 2 {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("no commands", func(t *testing.T) {
		if got := compactCmds(); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
	t.Run("only nil commands", func(t *testing.T) {
		if got := compactCmds(nil, nil); len(got) != 0 {
			t.Fatalf("got %v, want none", got)
		}
	})
}
