package cmd

// GetClipboard asks the terminal for the clipboard, which arrives as a ClipboardMsg.
type GetClipboard struct{}

func (GetClipboard) isCmd() {}

type SetClipboard []byte

func (SetClipboard) isCmd() {}
