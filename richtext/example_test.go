package richtext_test

import (
	"fmt"

	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
)

// A builder joins differently styled values into lines, such as a message after its author's name.
func ExampleBuilder() {
	var b richtext.Builder
	b.Write("alice", tcell.StyleDefault.Bold(true))
	b.Write(": hi\nhow are you?", tcell.StyleDefault)
	for _, line := range b.Finish() {
		for i, segment := range line {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%q bold=%v", segment.Text, segment.Style.HasBold())
		}
		fmt.Println()
	}
	// Output:
	// "alice" bold=true, ": hi" bold=false
	// "how are you?" bold=false
}

func ExampleWordWrap() {
	for _, line := range richtext.WordWrap("the quick brown fox jumps", 10) {
		fmt.Printf("%q\n", line)
	}
	// Output:
	// "the quick "
	// "brown fox "
	// "jumps"
}
