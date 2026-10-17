package tree_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/tree"
	"github.com/gdamore/tcell/v3"
)

type currentMsg struct{ node *tree.Node }

// A server's channels by category, hiding the root, with a row above each category and the voice channels collapsed. Down moves to the next node, which Update makes current.
func ExampleNew() {
	root := tree.NewNode("server")
	text := tree.NewNode("Text")
	text.AddChild(tree.NewNode("general")).AddChild(tree.NewNode("random"))
	voice := tree.NewNode("Voice").SetExpanded(false)
	voice.AddChild(tree.NewNode("lounge"))
	root.AddChild(text).AddChild(voice)

	channels := tree.New(root, text).
		TopLevel(1).
		Spacing(func(n *tree.Node) int {
			if len(n.Children()) > 0 {
				return 1
			}
			return 0
		}).
		Focused(true).
		OnChange(func(n *tree.Node) tview.Msg { return currentMsg{n} })
	screentest.Print(channels, 14, 5)

	msg := channels.Handle(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone), tview.Rectangle{Width: 14, Height: 5})
	fmt.Println(msg.(currentMsg).node.Spans()[0].Text)
	// Output:
	// ▾ Text
	// ├──general
	// └──random
	//
	// ▸ Voice
	// general
}
