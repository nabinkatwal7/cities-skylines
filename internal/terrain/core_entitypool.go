package terrain

type Entity interface {
	GetID() int
	SetID(id int)
	GetLifecycle() LifecycleState
	SetLifecycle(state LifecycleState)
}

type EntityPool[T Entity] struct {
	pool     []T
	freeList []int
	size     int
	capacity int
	factory  func() T
}

func NewEntityPool[T Entity](capacity int, factory func() T) *EntityPool[T] {
	p := &EntityPool[T]{
		pool:     make([]T, capacity),
		freeList: make([]int, capacity),
		capacity: capacity,
		factory:  factory,
	}
	for i := 0; i < capacity; i++ {
		p.pool[i] = factory()
		p.pool[i].SetID(-1)
		p.pool[i].SetLifecycle(LifecycleUnallocated)
		p.freeList[i] = i
	}
	return p
}

func (p *EntityPool[T]) Alloc(id int) int {
	if len(p.freeList) == 0 {
		p.grow()
	}
	idx := p.freeList[len(p.freeList)-1]
	p.freeList = p.freeList[:len(p.freeList)-1]
	e := p.pool[idx]
	e.SetID(id)
	e.SetLifecycle(LifecycleAllocated)
	p.size++
	return idx
}

func (p *EntityPool[T]) Free(index int) {
	if index < 0 || index >= p.capacity || p.pool[index].GetID() == -1 {
		return
	}
	e := p.pool[index]
	e.SetLifecycle(LifecycleReturnedToPool)
	e.SetID(-1)
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

func (p *EntityPool[T]) ForEach(fn func(T, int)) {
	for i := 0; i < p.capacity; i++ {
		e := p.pool[i]
		if e.GetID() == -1 {
			continue
		}
		lc := e.GetLifecycle()
		if lc == LifecycleActive || lc == LifecycleSuspended || lc == LifecycleMarkedForRemoval {
			fn(e, i)
		}
	}
}

func (p *EntityPool[T]) ForEachActive(fn func(T, int)) {
	for i := 0; i < p.capacity; i++ {
		e := p.pool[i]
		if e.GetLifecycle() == LifecycleActive {
			fn(e, i)
		}
	}
}

func (p *EntityPool[T]) Size() int     { return p.size }
func (p *EntityPool[T]) Capacity() int { return p.capacity }

func (p *EntityPool[T]) ToArray() []T {
	var result []T
	p.ForEach(func(e T, _ int) { result = append(result, e) })
	return result
}

func (p *EntityPool[T]) grow() {
	oldCap := p.capacity
	newCap := oldCap + max(oldCap/2, 64)
	newPool := make([]T, newCap)
	copy(newPool, p.pool)
	for i := oldCap; i < newCap; i++ {
		newPool[i] = p.factory()
		newPool[i].SetID(-1)
		newPool[i].SetLifecycle(LifecycleUnallocated)
		p.freeList = append(p.freeList, i)
	}
	p.pool = newPool
	p.capacity = newCap
}
