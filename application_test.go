package tview

import (
	"github.com/ayn2op/tview/cmd"
	"github.com/ayn2op/tview/layout"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/gdamore/tcell/v3/vt"
)

func TestAppendPasteKey(t *testing.T) {
	t.Parallel()

	var buffer strings.Builder
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, "b", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyCtrlJ, "", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, "c", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone),
	} {
		appendPasteKey(&buffer, event)
	}

	if got, want := buffer.String(), "a\nb\nc\t"; got != want {
		t.Fatalf("appendPasteKey() = %q, want %q", got, want)
	}
}

// cursorModel shows the cursor at x, y when drawn, unless x is negative.
type cursorModel struct {
	x, y int
}

var _ Model[cursorModel] = cursorModel{}

func (cursorModel) Init() Cmd                       { return nil }
func (m cursorModel) Update(Msg) (cursorModel, Cmd) { return m, nil }
func (m cursorModel) View() Widget                  { return m }
func (cursorModel) Handle(msg Msg, _ Rectangle) Msg { return msg }

func (m cursorModel) Draw(screen Screen, area Rectangle) {
	if m.x >= 0 {
		screen.ShowCursor(m.x, m.y)
	}
}

// labelModel draws its text at the top left.
type labelModel string

var _ Model[labelModel] = labelModel("")

func (labelModel) Init() Cmd                           { return nil }
func (m labelModel) Update(Msg) (labelModel, Cmd)      { return m, nil }
func (m labelModel) View() Widget                      { return m }
func (labelModel) Handle(msg Msg, _ Rectangle) Msg     { return msg }
func (labelModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }
func (labelModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}
func (m labelModel) Draw(screen Screen, area Rectangle) {
	screen.PutStr(area.X, area.Y, string(m))
}

func TestApplicationDraw(t *testing.T) {
	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 4, Y: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)

	t.Run("clears what an earlier frame drew", func(t *testing.T) {
		app := NewApplication(labelModel("long"), WithScreen(screen))
		app.draw()
		app.model = "x"
		app.draw()
		if str, _, _ := screen.Get(1, 0); str != " " {
			t.Fatalf("cell = %q, want it cleared", str)
		}
	})
	t.Run("hides a cursor from an earlier frame", func(t *testing.T) {
		screen.ShowCursor(1, 1)
		NewApplication(cursorModel{x: -1}, WithScreen(screen)).draw()
		if x, y := screen.GetCursor(); x != -1 || y != -1 {
			t.Fatalf("cursor at %d, %d, want hidden", x, y)
		}
	})
	t.Run("keeps a cursor shown in the frame", func(t *testing.T) {
		NewApplication(cursorModel{x: 2, y: 1}, WithScreen(screen)).draw()
		if x, y := screen.GetCursor(); x != 2 || y != 1 {
			t.Fatalf("cursor at %d, %d, want 2, 1", x, y)
		}
	})
}

// lengthModel's widget turns a string into its length, and Update quits on an int, keeping it.
type lengthModel struct {
	got *int
}

func (lengthModel) Init() Cmd { return cmd.Task(func() Msg { return "abc" }) }
func (m lengthModel) Update(msg Msg) (lengthModel, Cmd) {
	if n, ok := msg.(int); ok {
		*m.got = n
		return m, cmd.Quit{}
	}
	return m, nil
}
func (m lengthModel) View() Widget         { return m }
func (lengthModel) Draw(Screen, Rectangle) {}
func (lengthModel) Handle(msg Msg, _ Rectangle) Msg {
	if s, ok := msg.(string); ok {
		return len(s)
	}
	return msg
}

func TestApplicationRun(t *testing.T) {
	t.Run("passes messages from commands through the widget", func(t *testing.T) {
		screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 4, Y: 2}))
		if err != nil {
			t.Fatal(err)
		}
		if err := screen.Init(); err != nil {
			t.Fatal(err)
		}
		var got int
		if err := NewApplication(lengthModel{got: &got}, WithScreen(screen)).Run(); err != nil {
			t.Fatal(err)
		}
		if got != 3 {
			t.Fatalf("got %d, want 3", got)
		}
	})
	t.Run("sets colors in the order returned", func(t *testing.T) {
		tty := &recordingTty{MockTerm: vt.NewMockTerm(vt.MockOptSize{X: 4, Y: 2})}
		screen, err := tcell.NewTerminfoScreenFromTty(tty)
		if err != nil {
			t.Fatal(err)
		}
		if err := screen.Init(); err != nil {
			t.Fatal(err)
		}
		if err := NewApplication(colorModel{}, WithScreen(screen)).Run(); err != nil {
			t.Fatal(err)
		}
		written := tty.written.String()
		if set, reset := strings.Index(written, "\x1b]11;#000080"), strings.Index(written, "\x1b]111"); set < 0 || reset < set {
			t.Fatalf("background not set then restored: %q", written)
		}
	})
}

// recordingTty records what is written to the terminal.
type recordingTty struct {
	vt.MockTerm
	written strings.Builder
}

func (r *recordingTty) Write(b []byte) (int, error) {
	r.written.Write(b)
	return r.MockTerm.Write(b)
}

// colorModel sets the background, restores it, and quits.
type colorModel struct{ labelModel }

func (colorModel) Init() Cmd {
	return cmd.Batch{cmd.SetBackgroundColor(color.Navy), cmd.SetBackgroundColor(color.Default), cmd.Quit{}}
}
func (m colorModel) Update(Msg) (colorModel, Cmd) { return m, nil }

func (cursorModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (cursorModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

func (lengthModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (lengthModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// burstModel receives a burst of messages from commands and quits after the last, counting the frames drawn.
type burstModel struct {
	received, frames *int
}

const burstSize = 100

func (burstModel) Init() Cmd {
	cmds := make([]Cmd, burstSize)
	for i := range cmds {
		cmds[i] = cmd.Task(func() Msg { return i })
	}
	return cmd.Batch(cmds)
}

func (m burstModel) Update(msg Msg) (burstModel, Cmd) {
	if _, ok := msg.(int); ok {
		if *m.received++; *m.received == burstSize {
			return m, cmd.Quit{}
		}
	}
	return m, nil
}

func (m burstModel) View() Widget { return m }

// Draw takes as long as a frame of a busy screen.
func (m burstModel) Draw(Screen, Rectangle) {
	*m.frames++
	time.Sleep(time.Millisecond)
}
func (burstModel) Handle(msg Msg, _ Rectangle) Msg     { return msg }
func (burstModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }
func (burstModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// BenchmarkApplicationRunBurst handles a burst of messages from commands, as a screen loading many images at once does.
func BenchmarkApplicationRunBurst(b *testing.B) {
	var frames int
	for b.Loop() {
		screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 4, Y: 2}))
		if err != nil {
			b.Fatal(err)
		}
		if err := screen.Init(); err != nil {
			b.Fatal(err)
		}
		var received int
		if err := NewApplication(burstModel{received: &received, frames: &frames}, WithScreen(screen)).Run(); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(frames)/float64(b.N), "frames/op")
}
