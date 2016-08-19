package clikit

import (
	"context"

	"github.com/pkg/errors"
)

// CLI is a configured CLI.
type CLI struct {
	Root   Cmd
	Hooks  Hooks
	Parser Parser
}

// Hooks is the set of available hooks.
type Hooks struct {
	// PreParse is called before the command line is parsed. It is passed a
	// pointer to the raw args string which it may modify.
	PreParse func(cmdLine *[]string) error
	// PreExecute is called before the command is executed. It is passed the
	// same args pointer as PreParse, and the command that's about to be
	// executed.
	PreExecute func(Invocation) error
	// PostExecute is called after the command has executed. It is passed the
	// same args and Executor as PreExecute, and additionally a pointer to the
	// error returned by cmd.Execute(args). You can modify the error, e.g.
	// setting it to nil to ignore it.
	PostExecute func(Invocation, *error) error
}

// Invoke invokes the CLI to run a command.
func (c *CLI) Invoke(ctx context.Context, cmdLine []string) error {
	errChan := make(chan error)
	go func() {
		errChan <- c.invoke(ctx, cmdLine)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (c *CLI) invoke(ctx context.Context, cmdLine []string) error {
	if c.Hooks.PreParse != nil {
		if err := c.Hooks.PreParse(&cmdLine); err != nil {
			return errors.Wrap(err, "running pre-parse hook")
		}
	}
	invocation, parseErr := c.Parser.Parse(ctx, c.Root, cmdLine)
	if parseErr != nil {
		return errors.Wrap(parseErr, "parsing command line")
	}
	if c.Hooks.PreExecute != nil {
		if err := c.Hooks.PreExecute(invocation); err != nil {
			return errors.Wrap(err, "running pre-execute hook")
		}
	}
	err := invocation.Execute(ctx)
	if c.Hooks.PostExecute != nil {
		if err := c.Hooks.PostExecute(invocation, &err); err != nil {
			return errors.Wrap(err, "running post-execute hook")
		}
	}
	return err
}
