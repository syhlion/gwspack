package gwspack

import (
	"sync"
)

var (
	lock *sync.RWMutex   = new(sync.RWMutex)
	apps map[string]*app = make(map[string]*app)
)

func Get(key string) (c ClientController) {

	lock.Lock()
	defer lock.Unlock()
	if _, ok := apps[key]; !ok {
		apps[key] = newApp(key)
	}
	return apps[key]
}

func Info() (info map[string]int) {
	lock.RLock()
	defer lock.RUnlock()

	for k, v := range apps {
		info[k] = v.Count()
	}
	return info

}
