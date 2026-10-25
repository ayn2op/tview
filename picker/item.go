package picker

import "github.com/ayn2op/tview"

// Item is an entry of a picker, shown as its Content, such as text, and matched against the query by its FilterText.
type Item struct {
	Content    tview.Widget
	FilterText string
	Reference  any
}

type Items []Item

func (is Items) String(index int) string {
	return is[index].FilterText
}

func (is Items) Len() int {
	return len(is)
}
