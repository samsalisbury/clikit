package clikit

import (
	"context"
	"reflect"
)

// Invocation is a ready-to-execute invocation of a command.
type Invocation struct {
	Executer
	// Args are the non-flag args.
	Args []string
	// Options is a set of things collected by the command parser.
	Options OptionsSet
}

// Execute executes this invocation.
func (i Invocation) Execute(ctx context.Context) error {
	return i.Executer.Execute(ctx, i.Args)
}

// OptionsSet is a set of collected options structs.
type OptionsSet map[reflect.Type]interface{}

// Add adds a new struct.
func (optset *OptionsSet) Add(opts interface{}) {
	if *optset == nil {
		*optset = make(map[reflect.Type]interface{}, 1)
	}
	(*optset)[reflect.TypeOf(opts)] = opts
}

// AddAll adds all options from other. Where they clash, other wins.
func (optset *OptionsSet) AddAll(other *OptionsSet) {
	if *optset == nil {
		*optset = make(map[reflect.Type]interface{}, len(*other))
	}
	for k, v := range *other {
		(*optset)[k] = v
	}
}
