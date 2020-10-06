package main

import (
	"container/heap"
	"fmt"
	"time"
)

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*ExpTime

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].value.Before(pq[j].value)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

//Push new element to priority queue
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	expTime := x.(*ExpTime)
	expTime.index = n
	*pq = append(*pq, expTime)
}

//Pop remove root from queue and return it
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	expTime := old[n-1]
	old[n-1] = nil     // avoid memory leak
	expTime.index = -1 // for safety
	*pq = old[0 : n-1]
	return expTime
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(expTime *ExpTime, value time.Time) {
	expTime.value = value
	heap.Fix(pq, expTime.index)
}

//Peek return root value
func (pq PriorityQueue) Peek() *ExpTime {
	return pq[0]
}

func (pq PriorityQueue) String() string {
	var s string
	for i := 0; i < pq.Len(); i++ {
		expTime := fmt.Sprintf("index: %d; value: %v;\n", pq[i].index, pq[i].value)
		s = fmt.Sprint(s, expTime)
	}

	return s
}
