package injector

import (
	"sync"

	config "github.com/dynamicgo/go-config"
)

var globalInjector Injector
var globalContext Context
var globalOnce sync.Once

func globalInit() {
	globalInjector = New()
}

func globalContextInit() {
	globalContext = NewContext()
}

// RegisterService register service
func RegisterService(name string, f ServiceF) {
	globalOnce.Do(globalContextInit)

	globalContext.Register(name, f)
}

// RunServices run services
func RunServices(config config.Config) error {
	globalOnce.Do(globalContextInit)

	return globalContext.Run(config)
}

// Register call global injector with register function
func Register(key string, val interface{}) {
	globalOnce.Do(globalInit)

	globalInjector.Register(key, val)
}

// Get call global injector with get function
func Get(key string, val interface{}) bool {
	globalOnce.Do(globalInit)

	return globalInjector.Get(key, val)
}

// Find call global injector with Find function
func Find(val interface{}) {
	globalOnce.Do(globalInit)

	globalInjector.Find(val)
}

// Inject .
func Inject(val interface{}) error {
	globalOnce.Do(globalInit)

	return globalInjector.Inject(val)
}
