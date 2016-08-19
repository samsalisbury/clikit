package clikit

import (
	"context"
	"flag"
	"log"
	"reflect"
	"strings"
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
		fs := flag.NewFlagSet(name, flag.ContinueOnError)
		opts := &OptionsSet{}
		invocation, err = p.parse(fs, opts, root, name, cmdLine)
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

func (p *DefaultParser) parse(fs *flag.FlagSet, opts *OptionsSet, root Cmd, name string, cmdLine []string) (Invocation, error) {
	log.Printf("PARSING %T", root)
	if name[0] == byte('-') {
		panic("OMG")
		//return Invocation{}, errors.Errorf("command %q not recognised", name)
	}
	if executer, ok := root.(Executer); ok {
		log.Printf("EXECUTOR %T", root)
		if err := constructFlagSet(fs, opts, name, executer); err != nil {
			return Invocation{}, errors.Wrapf(err, "constructing flag set")
		}
		err := fs.Parse(cmdLine)
		return Invocation{
			Executer: executer,
			Args:     fs.Args(),
			Options:  *opts,
		}, err
	}
	var i Invocation
	if optioner, ok := root.(Optioner); ok {
		intermediateFS := flag.NewFlagSet(name, flag.ContinueOnError)
		intermediateOpts := optioner.Options()
		addFlags(intermediateFS, reflect.ValueOf(intermediateOpts))
		addFlags(fs, reflect.ValueOf(intermediateOpts))
		if err := intermediateFS.Parse(cmdLine); err != nil {
			return i, errors.Wrapf(err, "parsing intermediate options of %s", name)
		}
		opts.Add(intermediateOpts)
		cmdLine = intermediateFS.Args()
	}
	subcmdr, ok := root.(Subcmdr)
	if len(cmdLine) == 0 {
		return i, errors.Errorf("usage: %s %s", name, root.Help())
	}
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
	return p.parse(fs, opts, sub, subName, subCmdLine)
}

var optionSetType = reflect.TypeOf((*OptionSet)(nil)).Elem()

func constructFlagSet(fs1 *flag.FlagSet, opts *OptionsSet, name string, cmd Executer) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	t := reflect.TypeOf(cmd)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	addFlagGroupsFromType(fs, opts, t)
	return nil
}

// addFlagGroupsFromType adds fields of this type as flag groups.
func addFlagGroupsFromType(fs *flag.FlagSet, opts *OptionsSet, t reflect.Type) {
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
		if ft.Implements(optionSetType) || ftp.Implements(optionSetType) {
			addFlags(fs, fv)
			var optGroup interface{}
			if isPtr {
				optGroup = fv.Addr().Interface()
			} else {
				optGroup = fv.Interface()
			}
			opts.Add(optGroup)
			continue
		}
		addFlagGroupsFromType(fs, opts, ft)
	}
}

// addFlags adds the fields of this val as flags.
func addFlags(fs *flag.FlagSet, val reflect.Value) {
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}
	if typ.Kind() != reflect.Struct {
		panic("WANT STRUCT")
	}
	var infoFunc func(string) (interface{}, string, string)
	var ost OptionSet
	ptr := reflect.PtrTo(typ)
	if typ.Implements(optionSetType) {
		ost = reflect.New(typ).Elem().Interface().(OptionSet)
		infoFunc = ost.DefaultShortLong
	} else if ptr.Implements(optionSetType) {
		ost = reflect.New(typ).Interface().(OptionSet)
		infoFunc = ost.DefaultShortLong
	}

	n := typ.NumField()
	for i := 0; i < n; i++ {
		name := typ.Field(i).Name
		name = strings.ToLower(name)
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
