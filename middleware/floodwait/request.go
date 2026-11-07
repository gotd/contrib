package floodwait

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// object is a abstraction for Telegram API object with TypeID.
type object interface {
	TypeID() uint32
}

// queryWrapper is implemented by invokeWith* and invokeAfterMsg wrappers.
type queryWrapper interface {
	GetQuery() bin.Object
}

// maxUnwrap bounds nested wrappers unwrapping.
const maxUnwrap = 8

// unwrap returns innermost query of invokeWith* wrappers, e.g. the method
// called in a takeout session via invokeWithTakeout.
func unwrap(encoder bin.Encoder) bin.Encoder {
	for range maxUnwrap {
		w, ok := encoder.(queryWrapper)
		if !ok {
			return encoder
		}
		q := w.GetQuery()
		if q == nil {
			return encoder
		}
		encoder = q
	}
	return encoder
}

type key uint64

func (k *key) fromEncoder(encoder bin.Encoder) {
	obj, ok := unwrap(encoder).(object)
	if !ok {
		return
	}
	*k = key(obj.TypeID())
}

// typeName returns TL name of the (unwrapped) request, if known.
func typeName(encoder bin.Encoder) string {
	if v, ok := unwrap(encoder).(interface{ TypeName() string }); ok {
		return v.TypeName()
	}
	return ""
}

type request struct {
	ctx    context.Context
	input  bin.Encoder
	output bin.Decoder
	next   tg.Invoker
	key    key
	policy Policy

	retry       int
	serverRetry int
	result      chan error
}
