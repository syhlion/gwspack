package gwspack

import (
	"sync"
)

type connpool struct {
	lock *sync.RWMutex
	pool map[string]map[*client]UserData
}

func (cp *connpool) join(c *client) (err error) {

	cp.lock.Lock()
	defer cp.lock.Unlock()
	if v, ok := cp.pool[c.id]; !ok {
		m := make(map[*client]UserData)
		m[c] = c.data
		cp.pool[c.id] = m
	} else {
		v[c] = c.data
	}
	return
}

func (cp *connpool) remove(c *client) (err error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if _, ok := cp.pool[c.id]; ok {
		delete(cp.pool[c.id], c)
	}
	return
}
func (cp *connpool) removeById(id string) (err error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if _, ok := cp.pool[id]; ok {
		delete(cp.pool, id)
	}
	return
}

func (cp *connpool) countById() (i int) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	i = len(cp.pool)
	return

}

func (cp *connpool) count() (i int) {
	cp.lock.RLock()
	defer cp.lock.RUnlock()

	for k, _ := range cp.pool {
		for _, _ = range cp.pool[k] {
			i++
		}
	}
	return i
}

func (cp *connpool) sendTo(id string, b []byte) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for c := range cp.pool[id] {
		c.send <- b
	}
	return

}

func (cp *connpool) sendAll(b []byte) {

	for _, clientMap := range cp.pool {
		for client := range clientMap {
			client.send <- b
		}
	}

}
