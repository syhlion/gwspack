package gwspack

import (
	"errors"
	"sync"
)

var (
	lock *sync.RWMutex   = new(sync.RWMutex)
	apps map[string]*app = make(map[string]*app)
)

func New(key string, r Receiver, clientSets int) (c ClientController) {

	lock.Lock()
	defer lock.Unlock()
	if _, ok := apps[key]; !ok {
		apps[key] = newApp(key, r, clientSets)
		go apps[key].Run()
	}
	return apps[key]
}

func Get(key string) (c ClientController, err error) {

	lock.RLock()
	defer lock.RUnlock()
	c, ok := apps[key]
	if !ok {
		err = errors.New("empty")
	}

	return

}
