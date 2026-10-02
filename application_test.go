package tview

import (
	"github.com/ayn2op/tview/layout"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
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

// cursorScreen records where the cursor was last shown, with -1, -1 for hidden.
type cursorScreen struct {
	Screen
	x, y int
}

func (s *cursorScreen) ShowCursor(x, y int) { s.x, s.y = x, y }
func (s *cursorScreen) HideCursor()         { s.ShowCursor(-1, -1) }

// cursorModel shows the cursor at x, y when drawn, unless x is negative.
type cursorModel struct {
	x, y int
}

var _ Model[cursorModel] = cursorModel{}

func (cursorModel) Init() Cmd                       { return nil }
func (m cursorModel) Update(Msg) (cursorModel, Cmd) { return m, nil }
func (m cursorModel) View() Element                 { return m }
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
func (m labelModel) View() Element                     { return m }
func (labelModel) Handle(msg Msg, _ Rectangle) Msg     { return msg }
func (labelModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }
func (labelModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}
func (m labelModel) Draw(screen Screen, area Rectangle) {
	screen.PutStr(area.X, area.Y, string(m))
}

func TestApplicationDraw(t *testing.T) {
	mock, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 4, Y: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mock.Fini)
	screen := &cursorScreen{Screen: mock}

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
		if screen.x != -1 || screen.y != -1 {
			t.Fatalf("cursor at %d, %d, want hidden", screen.x, screen.y)
		}
	})
	t.Run("keeps a cursor shown in the frame", func(t *testing.T) {
		NewApplication(cursorModel{x: 2, y: 1}, WithScreen(screen)).draw()
		if screen.x != 2 || screen.y != 1 {
			t.Fatalf("cursor at %d, %d, want 2, 1", screen.x, screen.y)
		}
	})
}

// lengthModel's element turns a string into its length, and Update quits on an int, keeping it.
type lengthModel struct {
	got *int
}

func (lengthModel) Init() Cmd { return func() Msg { return "abc" } }
func (m lengthModel) Update(msg Msg) (lengthModel, Cmd) {
	if n, ok := msg.(int); ok {
		*m.got = n
		return m, Quit()
	}
	return m, nil
}
func (m lengthModel) View() Element        { return m }
func (lengthModel) Draw(Screen, Rectangle) {}
func (lengthModel) Handle(msg Msg, _ Rectangle) Msg {
	if s, ok := msg.(string); ok {
		return len(s)
	}
	return msg
}

func TestApplicationRun(t *testing.T) {
	t.Run("passes messages from commands through the element", func(t *testing.T) {
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
}

func (cursorModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (cursorModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

func (lengthModel) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (lengthModel) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}
