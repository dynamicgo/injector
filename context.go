package injector

import (
	"fmt"
	"reflect"
	"sort"
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
	Start() error
}

// PriorityRunnable .
type PriorityRunnable interface {
	Runnable
	Priority() int
}

// RunnableJoinable .
type RunnableJoinable interface {
	Runnable
	Join()
}

// Context .
type Context interface {
	Register(name string, F ServiceF)
	Bind(config config.Config) error
	Start() error
	Injector() Injector
	Join()
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
		context.DebugF("inject service %s with type %s", name, reflect.TypeOf(service).String())

		err := context.injector.Inject(service)

		if err != nil {
			return xerrors.Wrapf(err, "inject service %s with type %s -- failed", name, reflect.TypeOf(service).String())
		}

		runnable, ok := service.(Runnable)

		if ok {
			runnables[name] = runnable
		}
	}

	return nil
}

type namedRunnable struct {
	Runnable
	Name string
}

func (context *contextImpl) Join() {
	var wg sync.WaitGroup

	for name, runnable := range context.runnables {

		joinable, ok := runnable.(RunnableJoinable)

		if ok {
			wg.Add(1)
			context.DebugF("service %s started ...", name)
			context.doJoin(&wg, joinable)
			context.DebugF("service %s stopped", name)
		}
	}

	wg.Wait()
}

func (context *contextImpl) doJoin(wg *sync.WaitGroup, joinable RunnableJoinable) {
	defer wg.Done()

	joinable.Join()
}

func (context *contextImpl) Start() error {

	var prior []*namedRunnable
	var runnables []*namedRunnable

	for name, runnable := range context.runnables {

		if _, ok := runnable.(PriorityRunnable); ok {
			prior = append(prior, &namedRunnable{
				Runnable: runnable,
				Name:     name,
			})
		} else {
			runnables = append(runnables, &namedRunnable{
				Runnable: runnable,
				Name:     name,
			})
		}

		// context.DebugF("service %s started ...", name)

		// if err := runnable.Start(); err != nil {
		// 	context.ErrorF("service %s stopped with err: %s", name, err)
		// 	return err
		// }

		// context.DebugF("service %s stopped", name)
	}

	sort.Slice(prior, func(i, j int) bool {
		return prior[i].Runnable.(PriorityRunnable).Priority() < prior[j].Runnable.(PriorityRunnable).Priority()
	})

	for _, n := range prior {
		name := n.Name

		runnable := n.Runnable

		context.DebugF("service %s start ...", name)

		if err := runnable.Start(); err != nil {
			context.ErrorF("service %s stopped with err: %s", name, err)
			return err
		}

		context.DebugF("service %s started", name)
	}

	for _, n := range runnables {
		name := n.Name

		runnable := n.Runnable

		context.DebugF("service %s start ...", name)

		if err := runnable.Start(); err != nil {
			context.ErrorF("service %s stopped with err: %s", name, err)
			return err
		}

		context.DebugF("service %s started", name)
	}

	return nil
}
