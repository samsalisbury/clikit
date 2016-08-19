package clikit

import (
	"context"
	"flag"
	"log"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

// DefaultParser is the default parser implementation.
type DefaultParser struct{}

// Parse implements Parser.Parse using the flag package alongside Options to
// parse the command.
func (p *DefaultParser) Parse(ctx context.Context, root Cmd, cmdLine []string) (Invocation, error) {

	done := make(chan error)
	var err error
	invocation := Invocation{}
	go func() {
		name := cmdLine[0]
		cmdLine = cmdLine[1:]
		invocation, err = p.parse(root, name, cmdLine)
		done <- err
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-done:
		break
	}
	return invocation, err
}

func (p *DefaultParser) parse(root Cmd, name string, cmdLine []string) (Invocation, error) {
	log.Printf("PARSING %T", root)
	if name[0] == byte('-') {
		panic("OMG")
		//return Invocation{}, errors.Errorf("command %q not recognised", name)
	}
	if executer, ok := root.(Executer); ok {
		log.Printf("EXECUTOR %T", root)
		flagSet, flagValues := constructFlagSet(name, executer)
		err := flagSet.Parse(cmdLine)
		return Invocation{
			Executer: executer,
			Args:     flagSet.Args(),
			Options:  flagValues,
		}, err
	}
	var i Invocation
	if len(cmdLine) == 0 {
		return i, errors.Errorf("usage: %s %s", name, root.Help())
	}
	subcmdr, ok := root.(Subcmdr)
	if !ok {
		return i, errors.Errorf("command %T has no subcommands", root)
	}
	subs := subcmdr.Subcmds()
	subName := cmdLine[0]
	subCmdLine := cmdLine[1:]
	sub, ok := subs[cmdLine[0]]
	if !ok {
		return i, errors.Errorf("command %q not recognised", name)
	}
	log.Println("DESCENDING")
	return p.parse(sub, subName, subCmdLine)
}

var optionSetType = reflect.TypeOf((*OptionSet)(nil)).Elem()

func constructFlagSet(name string, cmd Executer) (*flag.FlagSet, []interface{}) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	t := reflect.TypeOf(cmd)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fs, nil
	}
	flagGroups := addFlagGroupsFromType(fs, t)
	return fs, flagGroups
}

// addFlagGroupsFromType adds fields of this type as flag groups.
func addFlagGroupsFromType(fs *flag.FlagSet, t reflect.Type) []interface{} {
	groups := []interface{}{}
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i).Type
		isPtr := false
		if ft.Kind() == reflect.Ptr {
			isPtr = true
			ft = ft.Elem()
		}
		if ft.Kind() != reflect.Struct {
			continue
		}
		ftp := reflect.PtrTo(ft)
		fv := reflect.New(ft).Elem()
		var ost OptionSet
		if ft.Implements(optionSetType) {
			ost = reflect.New(ft).Elem().Interface().(OptionSet)
		} else if ftp.Implements(optionSetType) {
			ost = reflect.New(ft).Interface().(OptionSet)
		}

		if ft.Implements(optionSetType) || ftp.Implements(optionSetType) {
			addFlags(fs, ft, fv, ost.DefaultShortLong)
			var flagVal interface{}
			if isPtr {
				flagVal = fv.Addr().Interface()
			} else {
				flagVal = fv.Interface()
			}
			groups = append(groups, flagVal)
			continue
		}
		groups = append(groups, addFlagGroupsFromType(fs, ft)...)
	}
	return groups
}

// addFlags adds the fields of this val as flags.
func addFlags(fs *flag.FlagSet, typ reflect.Type, val reflect.Value,
	infoFunc func(string) (interface{}, string, string)) {
	if typ.Kind() != reflect.Struct {
		panic("WANT STRUCT")
	}
	n := typ.NumField()
	for i := 0; i < n; i++ {
		name := typ.Field(i).Name
		def, short, _ := infoFunc(name)
		switch v := val.Field(i).Interface().(type) {
		case bool:
			if def == nil {
				def = false
			}
			fs.BoolVar(&v, name, def.(bool), short)
		case string:
			if def == nil {
				def = ""
			}
			fs.StringVar(&v, name, def.(string), short)
		case int:
			if def == nil {
				def = 0
			}
			fs.IntVar(&v, name, def.(int), short)
		case time.Duration:
			if def == nil {
				def = 1 * time.Second
			}
			fs.DurationVar(&v, name, def.(time.Duration), short)
		}
	}
}
