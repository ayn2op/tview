package picker

type Item struct {
	Text       string
	FilterText string
	Reference  any
}

type Items []Item

func (is Items) String(index int) string {
	if is[index].FilterText != "" {
		return is[index].FilterText
	}
	return is[index].Text
}

func (is Items) Len() int {
	return len(is)
}
