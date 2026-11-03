package cmd

type Task func() Msg

func (Task) isCmd() {}
