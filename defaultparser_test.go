package clikit

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

var defaultParserTests = map[string]struct {
	ExpectedExecuter Executer
	ExpectedArgs     []string
	ExpectedErr      string
	ExpectedOpts     []interface{}
}{
	"cmd": {
		ExpectedErr: "usage: cmd <command>",
	},
	"cmd -debug": {
		ExpectedErr: "usage: cmd <command>",
		ExpectedOpts: []interface{}{
			RootOptions{Debug: true},
		},
	},
	"cmd -configfile somefile": {
		ExpectedErr: "usage: cmd <command>",
		ExpectedOpts: []interface{}{
			RootOptions{ConfigFile: "somefile"},
		},
	},
	"cmd list": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
	},
	"cmd -debug list": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
		ExpectedOpts: []interface{}{
			RootOptions{Debug: true},
		},
	},
	"cmd -configfile somefile list": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
		ExpectedOpts: []interface{}{
			RootOptions{ConfigFile: "somefile"},
		},
	},
	"cmd -configfile somefile -debug list": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
		ExpectedOpts: []interface{}{
			RootOptions{ConfigFile: "somefile", Debug: true},
		},
	},
	"cmd list -debug": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
		ExpectedOpts: []interface{}{
			RootOptions{Debug: true},
		},
	},
	"cmd list -configfile somefile": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
		ExpectedOpts: []interface{}{
			RootOptions{ConfigFile: "somefile"},
		},
	},
	"cmd list -configfile somefile -debug": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
		ExpectedOpts: []interface{}{
			RootOptions{ConfigFile: "somefile", Debug: true},
		},
	},
	"cmd list blah blah blah": {
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{"blah", "blah", "blah"},
	},
}

func TestDefaultParser(t *testing.T) {
	for cmd, test := range defaultParserTests {
		t.Log(cmd)
		test := test
		command := strings.Fields(cmd)
		//t.Run(cmd, func(t *testing.T) {
		dp := DefaultParser{}
		ctx := context.Background()
		invocation, err := dp.Parse(ctx, &Root{}, command)
		if err := compareErrors(err, test.ExpectedErr); err != nil {
			t.Errorf("%v", errors.Wrap(err, cmd))
		}
		if err := compareExecutors(invocation.Executer, test.ExpectedExecuter); err != nil {
			t.Error(errors.Wrap(err, cmd))
		}
		if err := compareArgs(invocation.Args, test.ExpectedArgs); err != nil {
			t.Error(errors.Wrap(err, cmd))
		}
		//})
	}
}

func compareArgs(actual, expected []string) error {
	if len(actual) != len(expected) {
		return errors.Errorf("got %d args (%v); want %d (%v)",
			len(actual), actual, len(expected), expected)
	}
	for i, expected := range expected {
		if actual[i] != expected {
			return errors.Errorf("got arg %q at pos %d; want %q", actual, i, expected)
		}
	}
	return nil
}

func compareExecutors(actual, expected Executer) error {
	actualType, expectedType := reflect.TypeOf(actual), reflect.TypeOf(expected)
	if actualType != expectedType {
		return errors.Errorf("got executor type %s; want %s", actualType, expectedType)
	}
	return nil
}

func compareErrors(actualErr error, expected string) error {
	if actualErr == nil && expected == "" {
		return nil
	}
	if actualErr == nil && expected != "" {
		return errors.Errorf("got nil; want error %q", expected)
	}
	if actualErr != nil && expected == "" {
		return actualErr
	}
	actual := actualErr.Error()
	if actual != expected {
		return errors.Errorf("got error %#q; want %#q", actual, expected)
	}
	return nil

}
