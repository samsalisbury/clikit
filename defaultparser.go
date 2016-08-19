package clikit

import "context"

// DefaultParser is the default parser implementation.
type DefaultParser struct{}

// Parse implements Parser.Parse using the flag package alongside Options to
// parse the command.
func (p *DefaultParser) Parse(ctx context.Context, root Cmd, cmdLine []string) (Executer, []string, error) {
	var (
		executer Executer
		args     []string
		err      error
	)

	done := make(chan error)
	go func() {
		executer, args, err = p.parse(root, cmdLine)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case err := <-done:
		return executer, args, err
	}
}

func (p *DefaultParser) parse(root Cmd, cmdLine []string) (Executer, []string, error) {
	return nil, nil, nil
}
