package tview

import (
	"github.com/gdamore/tcell/v2"
)

// page represents one page of a Pages object.
type page struct {
	name    string    // The page's name.
	item    Primitive // The page's primitive.
	resize  bool      // Whether or not to resize the page when it is drawn.
	visible bool      // Whether or not this page is visible.
}

// Pages is a container for other primitives laid out on top of each other,
// overlapping or not. It is often used as the application's root primitive. It
// allows to easily switch the visibility of the contained primitives.
//
// See https://github.com/rivo/tview/wiki/Pages for an example.
type Pages struct {
	*Box

	// The contained pages. (Visible) pages are drawn from back to front.
	pages []*page

	// We keep a reference to the function which allows us to set the focus to
	// a newly visible page.
	setFocus func(p Primitive)

	// An optional handler which is called whenever the visibility or the order of
	// pages changes.
	changed func()
}

// NewPages returns a new Pages object.
func NewPages() *Pages {
	p := &Pages{
		Box: NewBox(),
	}
	return p
}

// SetChangedFunc sets a handler which is called whenever the visibility or the
// order of any visible pages changes. This can be used to redraw the pages.
func (p *Pages) SetChangedFunc(handler func()) *Pages {
	p.changed = handler
	return p
}

// GetPageCount returns the number of pages currently stored in this object.
func (p *Pages) GetPageCount() int {
	return len(p.pages)
}

// GetPageNames returns all page names ordered from front to back,
// optionally limited to visible pages.
func (p *Pages) GetPageNames(visibleOnly bool) []string {
	var names []string
	for index := len(p.pages) - 1; index >= 0; index-- {
		if !visibleOnly || p.pages[index].visible {
			names = append(names, p.pages[index].name)
		}
	}
	return names
}

func (p *Pages) GetVisibile(name string) bool {
	for _, page := range p.pages {
		if name == page.name {
			return page.visible
		}
	}

	return false
}

func (p *Pages) ClearPages() *Pages {
	p.pages = nil
	return p
}

// AddPage adds a new page with the given name and primitive. If there was
// previously a page with the same name, it is overwritten. Leaving the name
// empty may cause conflicts in other functions so always specify a non-empty
// name.
//
// Visible pages will be drawn in the order they were added (unless that order
// was changed in one of the other functions). If "resize" is set to true, the
// primitive will be set to the size available to the Pages primitive whenever
// the pages are drawn.
func (p *Pages) AddPage(name string, item Primitive, resize, visible bool) *Pages {
	hasFocus := p.HasFocus()
	for index, pg := range p.pages {
		if pg.name == name {
			p.pages = append(p.pages[:index], p.pages[index+1:]...)
			break
		}
	}
	p.pages = append(p.pages, &page{item: item, name: name, resize: resize, visible: visible})
	if p.changed != nil {
		p.changed()
	}
	if hasFocus {
		p.Focus(p.setFocus)
	}
	return p
}

// AddAndSwitchToPage calls AddPage(), then SwitchToPage() on that newly added
// page.
func (p *Pages) AddAndSwitchToPage(name string, item Primitive, resize bool) *Pages {
	p.AddPage(name, item, resize, true)
	p.SwitchToPage(name)
	return p
}

// RemovePage removes the page with the given name. If that page was the only
// visible page, visibility is assigned to the last page.
func (p *Pages) RemovePage(name string) *Pages {
	var isVisible bool
	hasFocus := p.HasFocus()
	for index, page := range p.pages {
		if page.name == name {
			isVisible = page.visible
			p.pages = append(p.pages[:index], p.pages[index+1:]...)
			if page.visible && p.changed != nil {
				p.changed()
			}
			break
		}
	}
	if isVisible {
		for index, page := range p.pages {
			if index < len(p.pages)-1 {
				if page.visible {
					break // There is a remaining visible page.
				}
			} else {
				page.visible = true // We need at least one visible page.
			}
		}
	}
	if hasFocus {
		p.Focus(p.setFocus)
	}
	return p
}

// HasPage returns true if a page with the given name exists in this object.
func (p *Pages) HasPage(name string) bool {
	for _, page := range p.pages {
		if page.name == name {
			return true
		}
	}
	return false
}

// ShowPage sets a page's visibility to "true" (in addition to any other pages
// which are already visible).
func (p *Pages) ShowPage(name string) *Pages {
	for _, page := range p.pages {
		if page.name == name {
			page.visible = true
			if p.changed != nil {
				p.changed()
			}
			break
		}
	}
	if p.HasFocus() {
		p.Focus(p.setFocus)
	}
	return p
}

// HidePage sets a page's visibility to "false".
func (p *Pages) HidePage(name string) *Pages {
	for _, page := range p.pages {
		if page.name == name {
			page.visible = false
			if p.changed != nil {
				p.changed()
			}
			break
		}
	}
	if p.HasFocus() {
		p.Focus(p.setFocus)
	}
	return p
}

// SwitchToPage sets a page's visibility to "true" and all other pages'
// visibility to "false".
func (p *Pages) SwitchToPage(name string) *Pages {
	for _, page := range p.pages {
		if page.name == name {
			page.visible = true
		} else {
			page.visible = false
		}
	}
	if p.changed != nil {
		p.changed()
	}
	if p.HasFocus() {
		p.Focus(p.setFocus)
	}
	return p
}

// SendToFront changes the order of the pages such that the page with the given
// name comes last, causing it to be drawn last with the next update (if
// visible).
func (p *Pages) SendToFront(name string) *Pages {
	for index, page := range p.pages {
		if page.name == name {
			if index < len(p.pages)-1 {
				p.pages = append(append(p.pages[:index], p.pages[index+1:]...), page)
			}
			if page.visible && p.changed != nil {
				p.changed()
			}
			break
		}
	}
	if p.HasFocus() {
		p.Focus(p.setFocus)
	}
	return p
}

// SendToBack changes the order of the pages such that the page with the given
// name comes first, causing it to be drawn first with the next update (if
// visible).
func (p *Pages) SendToBack(name string) *Pages {
	for index, pg := range p.pages {
		if pg.name == name {
			if index > 0 {
				p.pages = append(append([]*page{pg}, p.pages[:index]...), p.pages[index+1:]...)
			}
			if pg.visible && p.changed != nil {
				p.changed()
			}
			break
		}
	}
	if p.HasFocus() {
		p.Focus(p.setFocus)
	}
	return p
}

// GetFrontPage returns the front-most visible page. If there are no visible
// pages, ("", nil) is returned.
func (p *Pages) GetFrontPage() (name string, item Primitive) {
	for index := len(p.pages) - 1; index >= 0; index-- {
		if p.pages[index].visible {
			return p.pages[index].name, p.pages[index].item
		}
	}
	return
}

// HasFocus returns whether or not this primitive has focus.
func (p *Pages) HasFocus() bool {
	for _, page := range p.pages {
		if page.item.HasFocus() {
			return true
		}
	}
	return p.Box.HasFocus()
}

// Focus is called by the application when the primitive receives focus.
func (p *Pages) Focus(delegate func(p Primitive)) {
	if delegate == nil {
		return // We cannot delegate so we cannot focus.
	}
	p.setFocus = delegate
	var topItem Primitive
	for _, page := range p.pages {
		if page.visible {
			topItem = page.item
		}
	}
	if topItem != nil {
		delegate(topItem)
	} else {
		p.Box.Focus(delegate)
	}
}

// Draw draws this primitive onto the screen.
func (p *Pages) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	for _, page := range p.pages {
		if !page.visible {
			continue
		}
		if page.resize {
			x, y, width, height := p.GetInnerRect()
			page.item.SetRect(x, y, width, height)
		}
		page.item.Draw(screen)
	}
}

// MouseHandler returns the mouse handler for this primitive.
func (p *Pages) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return p.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if !p.InRect(event.Position()) {
			return false, nil
		}

		// Pass mouse events along to the last visible page item that takes it.
		for index := len(p.pages) - 1; index >= 0; index-- {
			page := p.pages[index]
			if page.visible {
				consumed, capture = page.item.MouseHandler()(action, event, setFocus)
				if consumed {
					return
				}
			}
		}

		return
	})
}

// InputHandler returns the handler for this primitive.
func (p *Pages) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		for _, page := range p.pages {
			if page.item.HasFocus() {
				if handler := page.item.InputHandler(); handler != nil {
					handler(event, setFocus)
					return
				}
			}
		}
	})
}

// PasteHandler returns the handler for this primitive.
func (p *Pages) PasteHandler() func(pastedText string, setFocus func(p Primitive)) {
	return p.WrapPasteHandler(func(pastedText string, setFocus func(p Primitive)) {
		for _, page := range p.pages {
			if page.item.HasFocus() {
				if handler := page.item.PasteHandler(); handler != nil {
					handler(pastedText, setFocus)
					return
				}
			}
		}
	})
}
