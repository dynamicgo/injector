package injector

import (
	"fmt"
	"sync"

	"github.com/dynamicgo/slf4go"

	config "github.com/dynamicgo/go-config"
)

// ServiceF .
type ServiceF func(config config.Config) (interface{}, error)

// Runnable .
type Runnable interface {
	Run() error
}

// Context .
type Context interface {
	Register(name string, F ServiceF)
	Run(config config.Config) error
}

type contextImpl struct {
	slf4go.Logger                     // logger .
	sync.RWMutex                      // mixin mutex
	injector      Injector            // injector instance
	services      map[string]ServiceF // services factories
}

// NewContext .
func NewContext() Context {
	return &contextImpl{
		Logger:   slf4go.Get("context"),
		injector: New(),
		services: make(map[string]ServiceF),
	}
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

func (context *contextImpl) Run(config config.Config) error {

	context.Lock()
	defer context.Unlock()

	services := make(map[string]interface{})

	for name, f := range context.services {
		service, err := f(config)

		if err != nil {
			return err
		}

		context.injector.Register(name, service)

		services[name] = service
	}

	runnables := make(map[string]Runnable)

	for name, service := range services {
		context.injector.Inject(service)

		runnable, ok := service.(Runnable)

		if ok {
			runnables[name] = runnable
		}
	}

	var wg sync.WaitGroup

	for name, runnable := range runnables {
		wg.Add(1)

		context.runService(&wg, name, runnable)
	}

	return nil
}

func (context *contextImpl) runService(wg *sync.WaitGroup, name string, runnable Runnable) {
	defer wg.Done()

	context.ErrorF("service %s started ...", name)

	if err := runnable.Run(); err != nil {
		context.ErrorF("service %s stopped with err: %s", name, err)
		return
	}

	context.ErrorF("service %s stopped", name)
}
