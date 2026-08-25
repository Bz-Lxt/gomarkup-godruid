package pool

// connList is an intrusive doubly linked list of live logical connections.
// It is the lifecycle source of truth. Not safe for concurrent use.
type connList struct {
	head, tail *Node
	len        int
}

func (l *connList) Len() int { return l.len }

func (l *connList) PushBack(n *Node) {
	n.prev = l.tail
	n.next = nil
	if l.tail != nil {
		l.tail.next = n
	} else {
		l.head = n
	}
	l.tail = n
	l.len++
}

func (l *connList) Remove(n *Node) {
	if n == nil {
		return
	}
	if n.prev != nil {
		n.prev.next = n.next
	} else if l.head == n {
		l.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else if l.tail == n {
		l.tail = n.prev
	}
	n.prev = nil
	n.next = nil
	if l.len > 0 {
		l.len--
	}
}

func (l *connList) Snapshot() []*Node {
	out := make([]*Node, 0, l.len)
	for n := l.head; n != nil; n = n.next {
		out = append(out, n)
	}
	return out
}

func (l *connList) ForEach(fn func(*Node) bool) {
	for n := l.head; n != nil; n = n.next {
		if !fn(n) {
			return
		}
	}
}
