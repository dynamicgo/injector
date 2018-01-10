package injector

import (
	"reflect"
	"strings"

	"github.com/dynamicgo/contract"
)

// Executor some type's inject executor
type Executor struct {
	Type reflect.Type // inject target type
	Ops  []Op         // execute op
}

// Context .
type Context interface {
	Get(target reflect.Type) (*reflect.Value, error)
	GetByPlaceHold(index int) (*reflect.Value, error)
}

// Op .
type Op interface {
	Execute(value *reflect.Value, context Context) error
}

// NewStructExecutor create new struct executor
func NewStructExecutor(target reflect.Type) (*Executor, error) {

	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	status := contract.Require(target.Kind() == reflect.Struct, "input target type must be struct")

	if !status.Ok {
		return nil, status
	}

	var ops []Op

	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		op, err := createStructFiledOp(&field)

		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}

	return &Executor{
		Type: target,
		Ops:  ops,
	}, nil
}

type injectFieldOp struct {
	nullable bool
	Type     reflect.Type
}

func (op *injectFieldOp) Execute(value *reflect.Value, context Context) error {
	return nil
}

func createStructFiledOp(field *reflect.StructField) (Op, error) {
	if value, ok := field.Tag.Lookup("inject"); ok {
		nullable := false

		if strings.Contains(value, "nullable") {
			nullable = true
		}

		return &injectFieldOp{
			nullable: nullable,
			Type:     field.Type,
		}, nil
	}

	return nil, nil
}

type bindArgsOp struct {
	index int
	Type  reflect.Type
}

// func createBindArgsOp(index int)
