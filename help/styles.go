package help

import (
	"github.com/gdamore/tcell/v3"
)

type Styles struct {
	ShortKey  tcell.Style
	ShortDesc tcell.Style

	FullKey  tcell.Style
	FullDesc tcell.Style

	ShortSeparator tcell.Style
	FullSeparator  tcell.Style
	Ellipsis       tcell.Style
}

func DefaultStyles() Styles {
	dim := tcell.StyleDefault.Dim(true)
	normal := tcell.StyleDefault
	return Styles{
		ShortKey:       dim,
		ShortDesc:      normal,
		ShortSeparator: dim,
		FullKey:        dim,
		FullDesc:       normal,
		FullSeparator:  dim,
		Ellipsis:       dim,
	}
}
