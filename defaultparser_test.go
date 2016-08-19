package clikit

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/pkg/errors"
)

var defaultParserTests = []struct {
	Command          []string
	ExpectedExecuter Executer
	ExpectedArgs     []string
	ExpectedErr      string
}{
	{
		Command:     []string{"blah"},
		ExpectedErr: "usage: blah <command>",
	},
	{
		Command:          []string{"blah", "list"},
		ExpectedExecuter: &List{},
		ExpectedArgs:     []string{},
	},
}

func TestDefaultParser(t *testing.T) {
	for i, test := range defaultParserTests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dp := DefaultParser{}
			ctx := context.Background()
			executer, args, err := dp.Parse(ctx, Root{}, test.Command)
			if err := compareErrors(err, test.ExpectedErr); err != nil {
				t.Error(err)
			}
			if err := compareExecutors(executer, test.ExpectedExecuter); err != nil {
				t.Error(err)
			}
			if err := compareArgs(args, test.ExpectedArgs); err != nil {
				t.Error(err)
			}
		})
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
		return errors.Errorf("got error %q; want %q", actual, expected)
	}
	return nil

}
