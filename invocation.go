package clikit

import "context"

// Invocation is a ready-to-execute invocation of a command.
type Invocation struct {
	Executer
	// Args are the non-flag args.
	Args []string
	// Options is a set of things collected by the command parser.
	Options []interface{}
}

// Execute executes this invocation.
func (i Invocation) Execute(ctx context.Context) error {
	return i.Executer.Execute(ctx, i.Args)
}
