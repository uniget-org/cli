package cli

import "github.com/spf13/cobra"

type CommandWrapper struct {
	cobra.Command
}

type Command interface {
	init() error
}

type CommandList []Command

func NewCommandList() *CommandList {
	return &CommandList{}
}

func (cl *CommandList) AddCommand(cmd Command) {
	*cl = append(*cl, cmd)
}
