package clikit

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// Root is the test root command.
type Root struct {
	Options RootOptions
}

func (Root) Help() string {
	return "root help"
}

type RootOptions struct {
	Debug      bool
	ConfigFile string
}

func (RootOptions) DefaultShortLong(fieldName string) (def interface{}, short, long string) {
	if fieldName == "Debug" {
		return false, "turn on debug level logging", ""
	}
	if fieldName == "ConfigFile" {
		return "~/.config/clikit/test/smalloptions", "configuration file path", ""
	}
	return nil, "", ""
}

func (Root) Subcmds() []Cmd {
	return []Cmd{
		&List{}, &Run{},
	}
}

type List struct {
	Options ListOptions
}

type ListOptions struct {
	JSON bool
}

func (ListOptions) DefaultShortLong(fieldName string) (def interface{}, short, long string) {
	return nil, "", ""
}

func (List) Help() string { return "list help" }

func (List) Execute(ctx context.Context, args []string) error {
	for i, s := range []string{"one", "two", "three"} {
		fmt.Println(i, "-", s)
	}
	return nil
}

type Run struct{}

func (Run) Help() string { return "run help" }

// ExampleOptionSet exemplifies a option set.
type ExampleOptionSet struct {
	Bool          bool
	String        string
	Duration      time.Duration
	Int           int
	FlagOnly      string `cli:",flagonly"`      // Not read from env or file.
	Named         string `cli:"other"`          // Called "other".
	NamedFlagOnly string `cli:"named,flagonly"` // Flag only, called "renamed".
	Ignored       string `cli:"-"`              // Ignored by flag resolver.
}

// DefaultShortLong returns default value, short description and long
// description of the named option.
func (ExampleOptionSet) DefaultShortLong(name string) (def interface{}, short, long string) {
	switch name {
	default:
		return nil, "", ""
	case "Bool":
		return nil,
			`a boolean flag, its mere presence on the cli sets it to true`,
			`
			Bool is a boolean flag, this is its longer description.
			`
	case "String":
		return "some-default",
			"a simple string option",
			""
	case "Duration":
		return 3*time.Minute + 2*time.Second,
			"a duration as a string, e.g. 3m2s",
			"longer description"
	case "Int":
		return runtime.NumCPU(),
			"an integer",
			"defaults to runtime.NumCPU()"
	case "FlagOnly":
		return nil,
			"a string, only read from flags",
			"blah"
	case "Named":
		return nil,
			"a named option",
			""
	case "NamedFlagOnly":
		return nil, "a named option, only read from flags", ""
	case "Ignored":
		panic(`this will not happen as the field has the cli:"-" tag`)
	}
}

var cliTests = []struct {
	CommandLine, OutLines []string
	Err                   string
}{
	{
		CommandLine: []string{"cmdname", "command", "-bool", "-string", "blah", "subcommand"},
		OutLines: []string{
			"",
		},
	},
}

func TestCLI(t *testing.T) {
	t.Skip()
	for i, test := range cliTests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			cli := CLI{
				Root: &Root{},
				Hooks: Hooks{
					PreParse: func(ctx context.Context, cmdLine *[]string) error {
						return nil
					},
					PostParse: func(ctx context.Context, cmd Executer, args []string) error {
						return nil
					},
					PreExecute: func(ctx context.Context, cmd Executer, args []string) error {
						return nil
					},
					PostExecute: func(ctx context.Context, cmd Executer, args []string, err error) error {
						return nil
					},
				},
				Parser: &DefaultParser{},
			}
			ctx := context.Background()
			err := cli.Invoke(ctx, test.CommandLine)

			e := func(format string, a ...interface{}) error {
				err := errors.Errorf(format, a...)
				return errors.Wrapf(err, "executing %# q", strings.Join(test.CommandLine, " "))
			}

			if err != nil && test.Err == "" {
				t.Error(e("unexpected error %q", err))
			}
			if test.Err != "" {
				if err == nil {
					t.Error(e("got nil; want error %q", test.Err))
				} else if err.Error() != test.Err {
					t.Error(e("got error %q; want %q", err.Error(), test.Err))
				}
			}
			for range test.OutLines {
				// todo: check output
			}
		})
	}
}
