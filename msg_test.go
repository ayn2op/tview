package tview

import "testing"

func TestCompactCmds(t *testing.T) {
	first := func() Msg { return 1 }
	second := func() Msg { return 2 }
	cmds := []Cmd{nil, first, nil, second, nil}

	got := compactCmds(cmds...)
	if len(got) != 2 || got[0]() != 1 || got[1]() != 2 {
		t.Fatalf("got %v", got)
	}
	if compactCmds() != nil || len(compactCmds(nil, nil)) != 0 {
		t.Fatal("nil commands were not removed")
	}
}
