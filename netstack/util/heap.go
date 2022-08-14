package util

type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

func NewHeap[T any](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{
		data: []T{},
		less: less,
	}
}

func (h *Heap[T]) heapify() {
	for i := 0; i < len(h.data); i++ {
		h.siftDown(i)
	}
}

func (h *Heap[T]) siftDown(i int) {
	l := h.left(i)
	r := h.right(i)
	smallest := i
	if l < len(h.data) && h.less(h.data[l], h.data[smallest]) {
		smallest = l
	}
	if r < len(h.data) && h.less(h.data[r], h.data[smallest]) {
		smallest = r
	}
	if smallest != i {
		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		h.siftDown(smallest)
	}
}

func (h *Heap[T]) siftUp(i int) {
	p := h.parent(i)
	if p >= 0 && h.less(h.data[i], h.data[p]) {
		h.data[p], h.data[i] = h.data[i], h.data[p]
		h.siftUp(p)
	}
}

func (h *Heap[T]) left(i int) int {
	return 2*i + 1
}

func (h *Heap[T]) right(i int) int {
	return 2*i + 2
}

func (h *Heap[T]) parent(i int) int {
	return (i - 1) / 2
}

func (h *Heap[T]) Pop() T {
	v := h.data[0]
	h.data[0] = h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	h.heapify()
	return v
}

func (h *Heap[T]) Push(v T) {
	h.data = append(h.data, v)
	h.siftUp(len(h.data) - 1)
}

func (h *Heap[T]) Peek() T {
	return h.data[0]
}

func (h *Heap[T]) Len() int {
	return len(h.data)
}