package layers

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// layer represents one layer of a Layers object.
type layer struct {
	name    string          // The layer's name.
	item    tview.Primitive // The layer's primitive.
	resize  bool            // Whether or not to resize the layer when it is drawn.
	visible bool            // Whether or not this layer is visible.
	enabled bool            // Whether or not this layer can receive focus/input.
	overlay bool            // Whether this layer applies a background style to layers behind it.
}

// Layers is a container for other primitives laid out on top of each other.
// The layers are drawn from back to front and can optionally apply a
// background style to the layers behind them (typically used for modal dialogs).
type Layers struct {
	*tview.Box

	// The contained layers. (Visible) layers are drawn from back to front.
	layers []*layer
	// The style applied to layers behind the active overlay layer.
	backgroundLayerStyle tcell.Style

	// We keep a reference to the function which allows us to set the focus to
	// a newly visible layer.
	setFocus func(p tview.Primitive)
	// An optional handler which is called whenever the visibility or the order of
	// layers changes.
	changed func()
}

// Option configures a layer on Add.
type Option func(*layer)

// WithName sets the layer's name.
func WithName(name string) Option {
	return func(l *layer) {
		l.name = name
	}
}

// WithResize sets whether the layer is resized to the container's inner rect.
func WithResize(resize bool) Option {
	return func(l *layer) {
		l.resize = resize
	}
}

// WithVisible sets the initial visibility of the layer.
func WithVisible(visible bool) Option {
	return func(l *layer) {
		l.visible = visible
	}
}

// WithEnabled sets whether the layer can receive focus and input.
func WithEnabled(enabled bool) Option {
	return func(l *layer) {
		l.enabled = enabled
	}
}

// WithOverlay marks this layer as an overlay layer.
func WithOverlay() Option {
	return func(l *layer) {
		l.overlay = true
	}
}

// New returns a new Layers object.
func New() *Layers {
	l := &Layers{Box: tview.NewBox()}
	return l
}

// SetChangedFunc sets a handler which is called whenever the visibility or the
// order of any visible layers changes. This can be used to redraw the layers.
func (l *Layers) SetChangedFunc(handler func()) *Layers {
	l.changed = handler
	return l
}

// GetLayerCount returns the number of layers currently stored in this object.
func (l *Layers) GetLayerCount() int {
	return len(l.layers)
}

// GetLayerNames returns all layer names ordered from front to back,
// optionally limited to visible layers.
func (l *Layers) GetLayerNames(visibleOnly bool) []string {
	var names []string
	for index := len(l.layers) - 1; index >= 0; index-- {
		if !visibleOnly || l.layers[index].visible {
			names = append(names, l.layers[index].name)
		}
	}
	return names
}

// GetVisible returns whether the given layer is visible.
func (l *Layers) GetVisible(name string) bool {
	for _, layer := range l.layers {
		if name == layer.name {
			return layer.visible
		}
	}
	return false
}

// Clear removes all layers.
func (l *Layers) Clear() *Layers {
	if len(l.layers) > 0 {
		l.layers = nil
		l.MarkDirty()
	}
	return l
}

// AddLayer adds a new layer for the given primitive. Options can configure
// name, visibility, resize, overlay, and enabled state.
func (l *Layers) AddLayer(item tview.Primitive, opts ...Option) *Layers {
	hasFocus := l.HasFocus()
	newLayer := &layer{
		item:    item,
		visible: true,
		enabled: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(newLayer)
		}
	}
	if newLayer.name != "" {
		for index, layer := range l.layers {
			if layer.name == newLayer.name {
				l.layers = append(l.layers[:index], l.layers[index+1:]...)
				l.MarkDirty()
				break
			}
		}
	}
	l.layers = append(l.layers, newLayer)
	l.MarkDirty()
	if l.changed != nil {
		l.changed()
	}
	if hasFocus {
		l.Focus(l.setFocus)
	}
	return l
}

// RemoveLayer removes the layer with the given name.
func (l *Layers) RemoveLayer(name string) *Layers {
	hasFocus := l.HasFocus()
	for index, layer := range l.layers {
		if layer.name == name {
			l.layers = append(l.layers[:index], l.layers[index+1:]...)
			l.MarkDirty()
			if layer.visible && l.changed != nil {
				l.changed()
			}
			break
		}
	}
	if hasFocus {
		l.Focus(l.setFocus)
	}
	return l
}

// HasLayer returns true if a layer with the given name exists in this object.
func (l *Layers) HasLayer(name string) bool {
	for _, layer := range l.layers {
		if layer.name == name {
			return true
		}
	}
	return false
}

// ShowLayer sets a layer's visibility to "true" (in addition to any other layers
// which are already visible).
func (l *Layers) ShowLayer(name string) *Layers {
	for _, layer := range l.layers {
		if layer.name == name && !layer.visible {
			layer.visible = true
			l.MarkDirty()
			if l.changed != nil {
				l.changed()
			}
			break
		}
	}
	if l.HasFocus() {
		l.Focus(l.setFocus)
	}
	return l
}

// HideLayer sets a layer's visibility to "false".
func (l *Layers) HideLayer(name string) *Layers {
	for _, layer := range l.layers {
		if layer.name == name && layer.visible {
			layer.visible = false
			l.MarkDirty()
			if l.changed != nil {
				l.changed()
			}
			break
		}
	}
	if l.HasFocus() {
		l.Focus(l.setFocus)
	}
	return l
}

// SendToFront changes the order of the layers such that the layer with the given
// name comes last, causing it to be drawn last with the next update (if visible).
func (l *Layers) SendToFront(name string) *Layers {
	for index, layer := range l.layers {
		if layer.name == name {
			if index < len(l.layers)-1 {
				l.layers = append(append(l.layers[:index], l.layers[index+1:]...), layer)
				l.MarkDirty()
			}
			if layer.visible && l.changed != nil {
				l.changed()
			}
			break
		}
	}
	if l.HasFocus() {
		l.Focus(l.setFocus)
	}
	return l
}

// SendToBack changes the order of the layers such that the layer with the given
// name comes first, causing it to be drawn first with the next update (if
// visible).
func (l *Layers) SendToBack(name string) *Layers {
	for index, ly := range l.layers {
		if ly.name == name {
			if index > 0 {
				l.layers = append(append([]*layer{ly}, l.layers[:index]...), l.layers[index+1:]...)
				l.MarkDirty()
			}
			if ly.visible && l.changed != nil {
				l.changed()
			}
			break
		}
	}
	if l.HasFocus() {
		l.Focus(l.setFocus)
	}
	return l
}

// GetFrontLayer returns the front-most visible layer. If there are no visible
// layers, ("", nil) is returned.
func (l *Layers) GetFrontLayer() (name string, item tview.Primitive) {
	for index := len(l.layers) - 1; index >= 0; index-- {
		if l.layers[index].visible {
			return l.layers[index].name, l.layers[index].item
		}
	}
	return
}

// GetLayer returns the layer with the given name. If no such layer exists, nil is
// returned.
func (l *Layers) GetLayer(name string) tview.Primitive {
	for _, layer := range l.layers {
		if layer.name == name {
			return layer.item
		}
	}
	return nil
}

// SetLayerEnabled enables or disables a layer. Disabled layers are still drawn
// (if visible) but do not receive focus or input.
func (l *Layers) SetLayerEnabled(name string, enabled bool) *Layers {
	hasFocus := l.HasFocus()
	for _, layer := range l.layers {
		if layer.name == name && layer.enabled != enabled {
			if !enabled && layer.item.HasFocus() {
				layer.item.Blur()
			}
			layer.enabled = enabled
			l.MarkDirty()
			if layer.visible && l.changed != nil {
				l.changed()
			}
			break
		}
	}
	if hasFocus {
		l.Focus(l.setFocus)
	}
	return l
}

// GetLayerEnabled returns whether the layer with the given name is enabled.
func (l *Layers) GetLayerEnabled(name string) bool {
	for _, layer := range l.layers {
		if layer.name == name {
			return layer.enabled
		}
	}
	return false
}

// ClearLayerOverlay disables overlay styling for the given layer.
func (l *Layers) ClearLayerOverlay(name string) *Layers {
	for _, layer := range l.layers {
		if layer.name == name && layer.overlay {
			layer.overlay = false
			l.MarkDirty()
			if layer.visible && l.changed != nil {
				l.changed()
			}
			break
		}
	}
	return l
}

// SetBackgroundLayerStyle sets the style applied to layers behind the active
// overlay layer.
func (l *Layers) SetBackgroundLayerStyle(style tcell.Style) *Layers {
	if l.backgroundLayerStyle != style {
		l.backgroundLayerStyle = style
		l.MarkDirty()
		if l.changed != nil {
			l.changed()
		}
	}
	return l
}

// IsDirty returns whether this primitive or one of its visible children needs redraw.
func (l *Layers) IsDirty() bool {
	if l.Box.IsDirty() {
		return true
	}
	for _, layer := range l.layers {
		if layer.visible && layer.item != nil && layer.item.IsDirty() {
			return true
		}
	}
	return false
}

// MarkClean marks this primitive and all children as clean.
func (l *Layers) MarkClean() {
	l.Box.MarkClean()
	for _, layer := range l.layers {
		if layer.item != nil {
			layer.item.MarkClean()
		}
	}
}

// HasFocus returns whether or not this primitive has focus.
func (l *Layers) HasFocus() bool {
	for _, layer := range l.layers {
		if layer.enabled && layer.item.HasFocus() {
			return true
		}
	}
	return l.Box.HasFocus()
}

// Focus is called by the application when the primitive receives focus.
func (l *Layers) Focus(delegate func(p tview.Primitive)) {
	if delegate == nil {
		return // We cannot delegate so we cannot focus.
	}
	l.setFocus = delegate
	if top := l.topVisibleEnabledLayer(); top != nil {
		delegate(top.item)
		return
	}
	l.Box.Focus(delegate)
}

// Draw draws this primitive onto the screen.
func (l *Layers) Draw(screen tcell.Screen) {
	l.DrawForSubclass(screen, l)

	overlayIndex := l.topVisibleEnabledOverlayIndex()
	var ovScreen *overlayScreen
	if overlayIndex >= 0 {
		ovScreen = newOverlayScreen(screen, l.backgroundLayerStyle)
	}
	for index, layer := range l.layers {
		if !layer.visible {
			continue
		}
		layerScreen := screen
		if ovScreen != nil && index < overlayIndex {
			// Draw lower layers through the overlay screen so only the touched
			// cells get styled (avoids a full-screen pass).
			layerScreen = ovScreen
		}
		if layer.resize {
			x, y, width, height := l.GetInnerRect()
			layer.item.SetRect(x, y, width, height)
		}
		layer.item.Draw(layerScreen)
	}
}

// MouseHandler returns the mouse handler for this primitive.
func (l *Layers) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return l.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !l.InRect(event.Position()) {
			return false, nil
		}

		overlayIndex := l.topVisibleEnabledOverlayIndex()

		// Pass mouse events along to the front-most visible layer that takes it,
		// but never to layers behind an active overlay layer.
		for index := len(l.layers) - 1; index >= 0; index-- {
			layer := l.layers[index]
			if !layer.visible || !layer.enabled {
				continue
			}
			if overlayIndex >= 0 && index < overlayIndex {
				break
			}
			consumed, capture = layer.item.MouseHandler()(action, event, setFocus)
			if consumed {
				return
			}
		}

		// If an overlay layer is active, block input to layers behind it even if
		// the top layer didn't consume the event.
		if overlayIndex >= 0 {
			return true, nil
		}

		return
	})
}

// InputHandler returns the handler for this primitive.
func (l *Layers) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		for _, layer := range l.layers {
			if layer.enabled && layer.item.HasFocus() {
				if handler := layer.item.InputHandler(); handler != nil {
					handler(event, setFocus)
					return
				}
			}
		}
	})
}

// PasteHandler returns the handler for this primitive.
func (l *Layers) PasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	return l.WrapPasteHandler(func(pastedText string, setFocus func(p tview.Primitive)) {
		for _, layer := range l.layers {
			if layer.enabled && layer.item.HasFocus() {
				if handler := layer.item.PasteHandler(); handler != nil {
					handler(pastedText, setFocus)
					return
				}
			}
		}
	})
}

func (l *Layers) topVisibleEnabledLayer() *layer {
	for index := len(l.layers) - 1; index >= 0; index-- {
		layer := l.layers[index]
		if layer.visible && layer.enabled {
			return layer
		}
	}
	return nil
}

// topVisibleEnabledOverlayIndex returns the index of the top-most overlay
// layer that is both visible and enabled. This is used so only one overlay
// is applied at a time.
func (l *Layers) topVisibleEnabledOverlayIndex() int {
	for index := len(l.layers) - 1; index >= 0; index-- {
		layer := l.layers[index]
		if layer.visible && layer.enabled && layer.overlay {
			return index
		}
	}
	return -1
}

type overlayScreen struct {
	tcell.Screen
	overlay tcell.Style
}

func newOverlayScreen(screen tcell.Screen, overlay tcell.Style) *overlayScreen {
	return &overlayScreen{
		Screen:  screen,
		overlay: overlay,
	}
}

func (s *overlayScreen) SetContent(x int, y int, primary rune, combining []rune, style tcell.Style) {
	s.Screen.SetContent(x, y, primary, combining, applyBackgroundStyle(style, s.overlay))
}

func (s *overlayScreen) Put(x int, y int, str string, style tcell.Style) (string, int) {
	return s.Screen.Put(x, y, str, applyBackgroundStyle(style, s.overlay))
}

func (s *overlayScreen) PutStr(x int, y int, str string) {
	// Use StyleDefault so the screen's default style still applies, then overlay.
	s.Screen.PutStrStyled(x, y, str, applyBackgroundStyle(tcell.StyleDefault, s.overlay))
}

func (s *overlayScreen) PutStrStyled(x int, y int, str string, style tcell.Style) {
	s.Screen.PutStrStyled(x, y, str, applyBackgroundStyle(style, s.overlay))
}

func applyBackgroundStyle(base tcell.Style, overlay tcell.Style) tcell.Style {
	overlayFg := overlay.GetForeground()
	overlayBg := overlay.GetBackground()

	// Apply overlay foreground/background only when explicitly set. This avoids
	// forcing defaults that could unexpectedly replace existing content colors.
	if overlayFg != tcell.ColorDefault {
		base = base.Foreground(overlayFg)
	}
	if overlayBg != tcell.ColorDefault {
		base = base.Background(overlayBg)
	}

	// Apply overlay attributes additively so the overlay never removes existing
	// attributes (e.g. underline for links).
	if overlay.HasBold() {
		base = base.Bold(true)
	}
	if overlay.HasBlink() {
		base = base.Blink(true)
	}
	if overlay.HasDim() {
		base = base.Dim(true)
	}
	if overlay.HasItalic() {
		base = base.Italic(true)
	}
	if overlay.HasReverse() {
		base = base.Reverse(true)
	}
	if overlay.HasStrikeThrough() {
		base = base.StrikeThrough(true)
	}
	if overlay.HasUnderline() {
		base = base.Underline(true)
	}

	return base
}
