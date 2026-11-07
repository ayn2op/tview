package cmd

type Msg any

// Cmd is work done outside Update, such as I/O, whose result comes back as a message.
type Cmd interface{ isCmd() }

// WriteRaw writes data, such as escape sequences, to the terminal as is.
type WriteRaw string

func (WriteRaw) isCmd() {}

type Quit struct{}

func (Quit) isCmd() {}

// Suspend runs a function while the terminal is suspended and sends the message it returns, if any.
type Suspend func() Msg

func (Suspend) isCmd() {}

// RequestTerminalInfo asks for the terminal's name and version, which arrive as a TerminalInfoMsg.
type RequestTerminalInfo struct{}

func (RequestTerminalInfo) isCmd() {}

type Notify struct{ Title, Body string }

func (Notify) isCmd() {}
