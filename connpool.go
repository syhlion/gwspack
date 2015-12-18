package gwspack

import (
	"regexp"
	"sync"
)

type connpool struct {
	lock *sync.RWMutex
	pool map[string]map[*client]UserData
}

func (cp *connpool) join(c *client) (err error) {

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
	if _, ok := cp.pool[c.id]; ok {
		delete(cp.pool[c.id], c)
		if len(cp.pool[c.id]) == 0 {
			delete(cp.pool, c.id)
		}
	}
	return
}
func (cp *connpool) removeById(id string) (err error) {
	if _, ok := cp.pool[id]; ok {
		delete(cp.pool, id)
	}
	return
}

func (cp *connpool) CountById() (i int) {
	i = len(cp.pool)
	return

}

func (cp *connpool) Count() (i int) {

	for k, _ := range cp.pool {
		for _, _ = range cp.pool[k] {
			i++
		}
	}
	return i
}

func (cp *connpool) sendTo(id string, b []byte) {
	for c := range cp.pool[id] {
		select {
		case c.send <- b:
		default:
			close(c.send)
			delete(cp.pool[id], c)
		}
	}
	return

}

func (cp *connpool) sendAll(b []byte) {

	for _, clientMap := range cp.pool {
		for c := range clientMap {
			select {
			case c.send <- b:
			default:
				close(c.send)
				delete(clientMap, c)
			}
		}
	}

}

func (cp *connpool) List() (list map[string]UserData) {
	list = make(map[string]UserData)

	for k, v := range cp.pool {
		for c := range v {
			list[k] = c.data
			break
		}
	}
	return

}

func (cp *connpool) sendByRegex(vailed *regexp.Regexp, b []byte) {

	for k, clientMap := range cp.pool {
		if vailed.MatchString(k) {
			for client := range clientMap {
				client.send <- b
			}
		}
	}
}
