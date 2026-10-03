package tview

// BorderSet is the characters a border is drawn with.
type BorderSet struct {
	Top         string
	Bottom      string
	Left        string
	Right       string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	TopT        string
	BottomT     string
	LeftT       string
	RightT      string
}

func BorderSetHidden() BorderSet {
	return BorderSet{
		Top:         " ",
		Bottom:      " ",
		Left:        " ",
		Right:       " ",
		TopLeft:     " ",
		TopRight:    " ",
		BottomLeft:  " ",
		BottomRight: " ",
		TopT:        " ",
		BottomT:     " ",
		LeftT:       " ",
		RightT:      " ",
	}
}

func BorderSetPlain() BorderSet {
	return BorderSet{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		TopT:        "┬",
		BottomT:     "┴",
		LeftT:       "├",
		RightT:      "┤",
	}
}

func BorderSetRound() BorderSet {
	b := BorderSetPlain()
	b.TopLeft = "╭"
	b.TopRight = "╮"
	b.BottomLeft = "╰"
	b.BottomRight = "╯"
	return b
}

func BorderSetThick() BorderSet {
	return BorderSet{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
		TopT:        "┳",
		BottomT:     "┻",
		LeftT:       "┣",
		RightT:      "┫",
	}
}

func BorderSetDouble() BorderSet {
	return BorderSet{
		Top:         "═",
		Bottom:      "═",
		Left:        "║",
		Right:       "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		TopT:        "╦",
		BottomT:     "╩",
		LeftT:       "╠",
		RightT:      "╣",
	}
}

type Borders uint

const (
	BordersTop Borders = 1 << iota
	BordersBottom
	BordersLeft
	BordersRight

	BordersNone Borders = 0
	BordersAll  Borders = BordersTop | BordersBottom | BordersLeft | BordersRight
)

func (b Borders) Has(flag Borders) bool {
	return b&flag != 0
}
