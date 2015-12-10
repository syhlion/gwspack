package wsexchange

import (
	"testing"
)

func TestGet(t *testing.T) {

	appMap := NewAppMap()
	a := new(App)
	a.key = "test"
	appMap.Join(a)
	b, err := appMap.Get("test")
	if b != a {
		t.Error("Get Error")
	}
	b, err = appMap.Get("test2")
	if err == nil {
		t.Error("No Error")
	}

}

func TestJoin(t *testing.T) {

	appMap := NewAppMap()
	a := new(App)
	a.key = "test"

	appMap.Join(a)
	b, err := appMap.Get("test")
	if b != a {
		t.Error("Join Error")
	}
	err = appMap.Join(a)
	if err == nil {
		t.Error("No Error")
	}
}

func TestDelete(t *testing.T) {
	appMap := NewAppMap()
	a := new(App)
	a.key = "test"
	appMap.Join(a)
	appMap.Delete("test")
	_, err := appMap.Get("test")
	if err == nil {
		t.Error("No Error")
	}
}
