package mpris

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/prop"
)

func newProp(value interface{}, write bool, emitValue bool, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	var emitFlag prop.EmitType
	if emitValue {
		emitFlag = prop.EmitTrue
	} else {
		emitFlag = prop.EmitInvalidates
	}
	return &prop.Prop{
		Value:    value,
		Writable: write,
		Emit:     emitFlag,
		Callback: cb,
	}
}
