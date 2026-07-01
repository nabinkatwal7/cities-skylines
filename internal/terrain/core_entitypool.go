package terrain

type EntityPool[T any] struct {
	pool     []T
	freeList []int
	size     int
	capacity int
	factory  func() T
}

func NewEntityPool[T any](capacity int, factory func() T) *EntityPool[T] {
	p := &EntityPool[T]{
		pool:     make([]T, capacity),
		freeList: make([]int, capacity),
		capacity: capacity,
		factory:  factory,
	}
	for i := 0; i < capacity; i++ {
		p.pool[i] = factory()
		p.freeList[i] = i
	}
	return p
}

func (p *EntityPool[T]) Alloc() int {
	if len(p.freeList) == 0 {
		p.grow()
	}
	idx := p.freeList[len(p.freeList)-1]
	p.freeList = p.freeList[:len(p.freeList)-1]
	p.size++
	return idx
}

func (p *EntityPool[T]) Free(index int) {
	if index < 0 || index >= p.capacity {
		return
	}
	p.freeList = append(p.freeList, index)
	p.size--
}

func (p *EntityPool[T]) Get(index int) T {
	if index < 0 || index >= p.capacity {
		var zero T
		return zero
	}
	return p.pool[index]
}

func (p *EntityPool[T]) Set(index int, v T) {
	if index >= 0 && index < p.capacity {
		p.pool[index] = v
	}
}

func (p *EntityPool[T]) ForEach(fn func(T, int)) {
	for i := 0; i < p.capacity; i++ {
		e := p.pool[i]
		fn(e, i)
	}
}

func (p *EntityPool[T]) Size() int     { return p.size }
func (p *EntityPool[T]) Capacity() int { return p.capacity }

func (p *EntityPool[T]) Slice() []T {
	return p.pool[:p.capacity]
}

func (p *EntityPool[T]) grow() {
	oldCap := p.capacity
	newCap := oldCap + max(oldCap/2, 64)
	newPool := make([]T, newCap)
	copy(newPool, p.pool)
	for i := oldCap; i < newCap; i++ {
		newPool[i] = p.factory()
		p.freeList = append(p.freeList, i)
	}
	p.pool = newPool
	p.capacity = newCap
}
