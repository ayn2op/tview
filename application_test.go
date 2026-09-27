package tview

import (
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

func (*cursorModel) Init() Cmd                       { return nil }
func (*cursorModel) Update(Msg) Cmd                  { return nil }
func (m *cursorModel) View() Element                 { return m }
func (*cursorModel) Handle(msg Msg, _ Rectangle) Msg { return msg }

func (m *cursorModel) Draw(screen Screen, area Rectangle) {
	if m.x >= 0 {
		screen.ShowCursor(m.x, m.y)
	}
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

	t.Run("hides a cursor from an earlier frame", func(t *testing.T) {
		screen.ShowCursor(1, 1)
		NewApplication(&cursorModel{x: -1}, WithScreen(screen)).draw()
		if screen.x != -1 || screen.y != -1 {
			t.Fatalf("cursor at %d, %d, want hidden", screen.x, screen.y)
		}
	})
	t.Run("keeps a cursor shown in the frame", func(t *testing.T) {
		NewApplication(&cursorModel{x: 2, y: 1}, WithScreen(screen)).draw()
		if screen.x != 2 || screen.y != 1 {
			t.Fatalf("cursor at %d, %d, want 2, 1", screen.x, screen.y)
		}
	})
}

func TestApplicationRun(t *testing.T) {
	t.Run("nil root", func(t *testing.T) {
		if err := NewApplication(nil).Run(); err == nil {
			t.Fatal("Run returned no error")
		}
	})
}
