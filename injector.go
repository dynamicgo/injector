package injector

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/dynamicgo/slf4go"

	"github.com/dynamicgo/xerrors"
)

// public errors
var (
	ErrResource = errors.New("resource not found")
)

// Injector .
type Injector interface {
	Register(key string, val interface{})
	Get(key string, val interface{}) bool
	Find(val interface{})
	Inject(val interface{}) error
}

type typeInjector struct {
	valueT reflect.Type
	sync.Map
}

func newTypeInjector(valueT reflect.Type) *typeInjector {

	return &typeInjector{
		valueT: valueT,
	}
}

func (injector *typeInjector) Register(key string, val interface{}) {
	_, loaded := injector.LoadOrStore(key, val)

	if loaded {
		panic(fmt.Sprintf("register same type service %s with name %s", injector.valueT, key))
	}
}

func (injector *typeInjector) Get(key string, val interface{}) bool {

	valueT := reflect.TypeOf(val).Elem()

	if valueT != injector.valueT && valueT.Kind() == reflect.Interface && !injector.valueT.Implements(valueT) {

		panic(fmt.Sprintf("invalid input: %s %s", injector.valueT.String(), valueT.String()))
	}

	val2, ok := injector.Load(key)

	if !ok {
		return false
	}

	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(val2))

	return true
}

func (injector *typeInjector) Find(target interface{}) {

	targetSlice := reflect.ValueOf(target)

	var values []interface{}

	injector.Range(func(key, val interface{}) bool {
		values = append(values, val)
		return true
	})

	sliceValue := reflect.MakeSlice(reflect.SliceOf(injector.valueT), len(values), len(values))

	for i := 0; i < len(values); i++ {
		sliceValue.Index(i).Set(reflect.ValueOf(values[i]))
	}

	println("slice len", sliceValue.Len())

	targetSlice.Elem().Set(sliceValue)

}

type injectorImpl struct {
	sync.Map
	slf4go.Logger
}

// New .
func New() Injector {
	return &injectorImpl{
		Logger: slf4go.Get("inject"),
	}
}

func (inject *injectorImpl) getTypeInjector(valueT reflect.Type) *typeInjector {
	injectT, ok := inject.Load(valueT)

	if !ok {

		if valueT.Kind() == reflect.Interface {

			inject.Map.Range(func(key, val interface{}) bool {

				valT := key.(reflect.Type)

				if valT.Implements(valueT) {
					injectT = val.(*typeInjector)
					return false
				}

				return true
			})
		}

		if injectT == nil {
			injectT = newTypeInjector(valueT)

			injectT, _ = inject.LoadOrStore(valueT, injectT)
		}

	}

	return injectT.(*typeInjector)
}

func (inject *injectorImpl) Register(key string, val interface{}) {
	t := reflect.TypeOf(val)

	inject.DebugF("register service %p with type: %s", val, t.String())

	inject.getTypeInjector(t).Register(key, val)
}

func (inject *injectorImpl) Get(key string, val interface{}) bool {
	t := reflect.TypeOf(val)

	if t.Kind() != reflect.Ptr {
		panic("invalid input value, expect ptr of struct or interface")
	}

	t = t.Elem()

	if (t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct) && t.Kind() != reflect.Interface {
		panic("invalid inject field type, expect ptr of struct or interface")
	}

	return inject.getTypeInjector(t).Get(key, val)
}

func (inject *injectorImpl) Find(val interface{}) {
	t := reflect.TypeOf(val)

	if t.Kind() != reflect.Ptr && t.Elem().Kind() != reflect.Slice {
		panic("invalid input value,expect ptr of slice")
	}

	t = t.Elem().Elem()

	if (t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct) && t.Kind() != reflect.Interface {
		panic("invalid inject field type, expect ptr of struct or interface")
	}

	inject.getTypeInjector(t).Find(val)
}

// Inject .
func (inject *injectorImpl) Inject(target interface{}) error {

	valT := reflect.TypeOf(target)

	if valT.Kind() != reflect.Ptr || valT.Elem().Kind() != reflect.Struct {
		panic("inject target must be struct ptr")
	}

	structT := valT.Elem()
	structValue := reflect.ValueOf(target).Elem()

	for i := 0; i < structT.NumField(); i++ {

		field := structT.Field(i)

		tagStr, ok := field.Tag.Lookup("inject")

		if !ok {
			continue
		}

		if strings.ToTitle(field.Name[:1]) != field.Name[:1] {
			panic(fmt.Sprintf("inject filed must be export: %s", field.Name))
		}

		tag, err := inject.parseTag(tagStr)

		if err != nil {
			return err
		}

		if err := inject.executeInjectWithTag(tag, structValue.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

type injectTag struct {
	Name string
}

func (inject *injectorImpl) parseTag(tag string) (*injectTag, error) {
	return &injectTag{
		Name: tag,
	}, nil
}

func (inject *injectorImpl) executeInjectWithTag(tag *injectTag, fieldValue reflect.Value) error {
	fieldType := fieldValue.Type()

	if (fieldType.Kind() != reflect.Ptr || fieldType.Elem().Kind() != reflect.Struct) && fieldType.Kind() != reflect.Interface {
		panic("invalid inject field type, expect ptr of struct or interface")
	}

	if ok := inject.Get(tag.Name, fieldValue.Addr().Interface()); !ok {
		return xerrors.Wrapf(ErrResource, "inject object %s with type %s not found", tag.Name, fieldType)
	}

	return nil
}
