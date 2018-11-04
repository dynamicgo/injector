package injector

import (
	"errors"
	"sync"
)

// Errors
var (
	ErrNotFound = errors.New("resource not found")
)

// Injector injector engine
type Injector interface {
	Register(name string, service interface{})
	Get(name string, service interface{}) bool
	Find(services interface{}) // get services
	Bind(service interface{}) error
}

type injectorImpl struct {
	sync.RWMutex
	services map[string]interface{}
}

// New create new injector context
func New() Injector {
	return &injectorImpl{
		services: make(map[string]interface{}),
	}
}

func (injector *injectorImpl) Register(name string, service interface{}) {

}

func (injector *injectorImpl) Get(name string, service interface{}) bool {

}

func (injector *injectorImpl) Find(services interface{}) {

}

func (injector *injectorImpl) Bind(service interface{}) error {

}
