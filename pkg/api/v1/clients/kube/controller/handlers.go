package controller

import (
	"sync"

	"k8s.io/client-go/tools/cache"
)

// returns a handler that runs f() every time an update occurs,
// regardless of which type of update
func NewSyncHandler(f func()) cache.ResourceEventHandler {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			f()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			f()
		},
		DeleteFunc: func(obj interface{}) {
			f()
		},
	}
}

// returns a handler that runs f() every time an update occurs,
// regardless of which type of update
// ensures only one f() can run at a time
func NewLockingSyncHandler(f func()) cache.ResourceEventHandler {
	var mu sync.Mutex
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mu.Lock()
			f()
			mu.Unlock()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			mu.Lock()
			f()
			mu.Unlock()
		},
		DeleteFunc: func(obj interface{}) {
			mu.Lock()
			f()
			mu.Unlock()
		},
	}
}
