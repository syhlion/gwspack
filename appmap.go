package wsexchange

import (
	"errors"
	"sync"
)

type AppMap struct {
	lock *sync.RWMutex
	apps map[string]*App
}

func NewAppMap() *AppMap {
	return &AppMap{
		lock: new(sync.RWMutex),
		apps: make(map[string]*App),
	}
}

func (ac *AppMap) Join(a *App) (err error) {

	ac.lock.Lock()
	defer ac.lock.Unlock()
	if _, ok := ac.apps[a.key]; !ok {
		ac.apps[a.key] = a
	} else {
		err = errors.New("duplicate")
	}
	return
}

func (ac *AppMap) Get(key string) (app *App, err error) {
	ac.lock.RLock()
	defer ac.lock.RUnlock()
	if _, ok := ac.apps[key]; ok {
		app = ac.apps[key]
	} else {
		err = errors.New("empty")
	}
	return
}

func (ac *AppMap) Delete(key string) {
	ac.lock.Lock()
	defer ac.lock.Unlock()

	delete(ac.apps, key)
	return

}
