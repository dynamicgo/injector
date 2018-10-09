package injector

import (
	"fmt"
	"sync"

	"github.com/dynamicgo/xerrors"

	"github.com/dynamicgo/go-config-extend"

	"github.com/dynamicgo/slf4go"

	config "github.com/dynamicgo/go-config"
)

// Service .
type Service interface{}

// ServiceF .
type ServiceF func(config config.Config) (Service, error)

// Runnable .
type Runnable interface {
	Service
	Run() error
}

// Context .
type Context interface {
	Register(name string, F ServiceF)
	Bind(config config.Config) error
	Run() error
	Injector() Injector
}

type contextImpl struct {
	slf4go.Logger                     // logger .
	sync.RWMutex                      // mixin mutex
	injector      Injector            // injector instance
	services      map[string]ServiceF // services factories
	runnables     map[string]Runnable // runnable services
}

// NewContext .
func NewContext() Context {
	return &contextImpl{
		Logger:    slf4go.Get("context"),
		injector:  New(),
		services:  make(map[string]ServiceF),
		runnables: make(map[string]Runnable),
	}
}

func (context *contextImpl) Injector() Injector {
	return context.injector
}

func (context *contextImpl) Register(name string, f ServiceF) {
	context.Lock()
	defer context.Unlock()

	_, ok := context.services[name]

	if ok {
		panic(fmt.Sprintf("register same service: %s", name))
	}

	context.services[name] = f
}

func (context *contextImpl) Bind(config config.Config) error {
	context.Lock()
	defer context.Unlock()

	services := make(map[string]interface{})

	for name, f := range context.services {

		subconf, err := extend.SubConfig(config, "injector", name)

		if err != nil {
			return xerrors.Wrapf(err, "get service %s config error, %s", name, err)
		}

		service, err := f(subconf)

		if err != nil {
			return err
		}

		if service == nil {
			continue
		}

		context.injector.Register(name, service)

		services[name] = service
	}

	runnables := context.runnables

	for name, service := range services {
		context.injector.Inject(service)

		runnable, ok := service.(Runnable)

		if ok {
			runnables[name] = runnable
		}
	}

	return nil
}

func (context *contextImpl) Run() error {

	var wg sync.WaitGroup

	for name, runnable := range context.runnables {
		wg.Add(1)

		context.runService(&wg, name, runnable)
	}

	return nil
}

func (context *contextImpl) runService(wg *sync.WaitGroup, name string, runnable Runnable) {
	defer wg.Done()

	context.DebugF("service %s started ...", name)

	if err := runnable.Run(); err != nil {
		context.ErrorF("service %s stopped with err: %s", name, err)
		return
	}

	context.DebugF("service %s stopped", name)
}
