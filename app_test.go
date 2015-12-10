package wsexchange

import (
	"testing"
)

func TestIsExpand(t *testing.T) {
	var b bool
	b = IsExpand(500, 500, 0)
	if b != true {
		t.Error("Total:500, ClientSets:500, ProcessCount:0", b)
	}

	b = IsExpand(450, 500, 0)
	if b != true {
		t.Error("Total:450, ClientSets:500, ProcessCount:0", b)
	}
	b = IsExpand(600, 500, 1)
	if b != false {
		t.Error("Total:600, ClientSets:500, ProcessCount:1", b)
	}
	b = IsExpand(750, 500, 1)
	if b != false {
		t.Error("Total:750, ClientSets:500, ProcessCount:1", b)
	}
	b = IsExpand(756, 500, 1)
	if b != true {
		t.Error("Total:756, ClientSets:500, ProcessCount:1", b)
	}

}
func TestIsReduce(t *testing.T) {
	var b bool
	b = IsReduce(490, 500, 1)
	if b != false {
		t.Error("Total:490, ClientSets:500, ProcessCount:1", b)
	}
	b = IsReduce(1, 500, 1)
	if b != false {
		t.Error("Total:1, ClientSets:500, ProcessCount:1", b)
	}
	b = IsReduce(0, 500, 1)
	if b != false {
		t.Error("Total:0, ClientSets:500, ProcessCount:1", b)
	}
	b = IsReduce(760, 500, 2)
	if b != false {
		t.Error("Total:760, ClientSets:500, ProcessCount:2", b)
	}
	b = IsReduce(730, 500, 2)
	if b != true {
		t.Error("Total:730, ClientSets:500, ProcessCount:2", b)
	}

}
