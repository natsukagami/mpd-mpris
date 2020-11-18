package mpris

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

// Creates a new property, it is writable if `cb != nil`.
// If you need a writable prop without a handler, pass `notImplemented`.
func newProp(value interface{}, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	writable := cb == nil
	var emitFlag prop.EmitType
	if writable {
		emitFlag = prop.EmitTrue
	} else {
		emitFlag = prop.EmitFalse
	}
	return &prop.Prop{
		Value:    value,
		Writable: writable,
		Emit:     emitFlag,
		Callback: cb,
	}
}
