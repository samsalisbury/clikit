package clikit

import "context"

// Cmd is a command.
type Cmd interface {
	Helper
}

// Helper is a command with help. This is the only mandatory interface a command
// must implement.
type Helper interface {
	Help() string
}

// Subcmdr is a command with subcommands.
type Subcmdr interface {
	Subcmds() map[string]Cmd
}

// Executer is a command that can be directly executed. This means it can be the
// final command in an invocation.
type Executer interface {
	// Execute is passed a non-nil Context and any trailing non-flag arguments.
	Execute(ctx context.Context, args []string) error
}

// Optioner is a command, or a command dependency that accepts options.
type Optioner interface {
	Options() interface{}
}

// OptionSet is a set of options.
type OptionSet interface {
	// DefaultShortLong returns default value, short description and long
	// description of the named option.
	DefaultShortLong(fieldName string) (def interface{}, short, long string)
}

// Parser is something which can parse a command line to create an execution
// plan.
type Parser interface {
	// Parse parses a command line to return a target executer, args, and a
	// slice of values. The slice of values can, for example, be used to
	// represent the results of flag parsing.
	Parse(ctx context.Context, root Cmd, cmdLine []string) (Invocation, error)
}
