package gwspack

import (
	"regexp"
	"sync"
)

type connpool struct {
	lock              *sync.RWMutex
	pool              map[string]map[*client]UserData
	registerHandler   func(id string, s Sender, data UserData)
	unregisterHandler func(id string, s Sender, data UserData)
}

func (cp *connpool) Join(c *client) (err error) {

	cp.lock.Lock()
	defer cp.lock.Unlock()
	if v, ok := cp.pool[c.id]; !ok {
		m := make(map[*client]UserData)
		m[c] = c.data
		cp.pool[c.id] = m
		if cp.registerHandler != nil {
			cp.registerHandler(c.id, cp, c.data)
		}
	} else {
		v[c] = c.data
	}
	return
}

func (cp *connpool) Remove(c *client) (err error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if _, ok := cp.pool[c.id]; ok {
		delete(cp.pool[c.id], c)
		if cp.pool[c.id] == nil {
			if cp.unregisterHandler != nil {
				cp.unregisterHandler(c.id, cp, c.data)
			}
			delete(cp.pool, c.id)
		}
	}

	return
}
func (cp *connpool) RemoveById(id string) (err error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	var data UserData
	if m, ok := cp.pool[id]; ok {
		for _, v := range m {
			data = v
			break
		}
		if cp.unregisterHandler != nil {
			cp.unregisterHandler(id, cp, data)
		}
		delete(cp.pool, id)
	}
	return
}

func (cp *connpool) CountById() (i int) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	i = len(cp.pool)
	return

}

func (cp *connpool) Count() (i int) {
	cp.lock.RLock()
	defer cp.lock.RUnlock()

	for k, _ := range cp.pool {
		for _, _ = range cp.pool[k] {
			i++
		}
	}
	return i
}

func (cp *connpool) SendTo(id string, b []byte) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for c := range cp.pool[id] {
		c.send <- b
	}
	return

}

func (cp *connpool) SendAll(b []byte) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for _, clientMap := range cp.pool {
		for client := range clientMap {
			client.send <- b
		}
	}

}

func (cp *connpool) SendByRegex(regex string, b []byte) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for k, clientMap := range cp.pool {
		if vailed, err := regexp.Compile(regex); err == nil {
			if vailed.MatchString(k) {
				for client := range clientMap {
					client.send <- b
				}
			}
		}
	}
}
