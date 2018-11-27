package util

type IterativeQueue []interface{}

func NewIterativeQueue() IterativeQueue {
	iq := make(IterativeQueue, 0)
	return iq
}

func (iq *IterativeQueue) Enqueue(item interface{}) {
	*iq = append(*iq, item)
}

func (iq *IterativeQueue) Peek() interface{} {
	return (*iq)[0]
}

func (iq *IterativeQueue) Dequeue() interface{} {
	first := (*iq)[0]
	*iq = (*iq)[1:]
	return first
}

func (iq *IterativeQueue) Len() int {
	return len(*iq)
}
