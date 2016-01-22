package gwspack

import (
	"sync"
)

var (
	lock *sync.RWMutex   = new(sync.RWMutex)
	apps map[string]*app = make(map[string]*app)
)

// always return instance
func Get(key string) (c ClientController) {

	lock.Lock()
	defer lock.Unlock()
	if _, ok := apps[key]; !ok {
		apps[key] = newApp(key)
		go apps[key].run()
	}
	return apps[key]
}

// if exists or return nil
func Find(key string) ClientController {
	lock.Lock()
	defer lock.Unlock()

	if c, ok := apps[key]; ok {
		return c
	}

	return nil
}

func Info() (info map[string]int) {
	lock.RLock()
	defer lock.RUnlock()

	info = make(map[string]int)
	for k, v := range apps {
		info[k] = v.Count()
	}
	return info

}

func ClientList(key string) map[string]UserData {

	if c := Find(key); c != nil {
		return c.List()
	}

	return nil
}
