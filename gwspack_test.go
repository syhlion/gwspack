package gwspack

import "testing"

func TestFind(t *testing.T) {
	if c := Find("a"); c != nil {
		t.Errorf("Find(a) must be nil!! But %#v", c)
	}

	Get("b")

	if c := Find("b"); c == nil {
		t.Error("Find(b) after Create must be ok!!")
	}
}

func TestInfo(t *testing.T) {
	Get("a")
	Get("b")
	Get("c")

	data := []string{"a", "b", "c"}
	info := Info()
	for _, k := range data {
		if _, ok := info[k]; !ok {
			t.Error("Info() must contain [%s]", k)
		}
	}
}

func TestClientList(t *testing.T) {
	// 這個不好測...
	t.Log(ClientList("b"))
}
