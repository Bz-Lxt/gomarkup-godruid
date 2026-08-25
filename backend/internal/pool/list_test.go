package pool

import "testing"

func TestConnList(t *testing.T) {
	var l connList
	a, b, c := &Node{id: "a"}, &Node{id: "b"}, &Node{id: "c"}
	l.PushBack(a)
	l.PushBack(b)
	l.PushBack(c)
	if l.Len() != 3 {
		t.Fatalf("len %d", l.Len())
	}
	l.Remove(b)
	ids := []string{}
	l.ForEach(func(n *Node) bool {
		ids = append(ids, n.id)
		return true
	})
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "c" {
		t.Fatalf("ids %v", ids)
	}
	l.Remove(a)
	l.Remove(c)
	if l.Len() != 0 || l.head != nil || l.tail != nil {
		t.Fatal("list not empty")
	}
}
