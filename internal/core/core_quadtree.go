package core

type SpatialBounds struct {
	X, Y, W, H float64
}

type SpatialEntity interface {
	GetX() float64
	GetY() float64
	GetID() int
}

type QuadTree struct {
	bounds   SpatialBounds
	capacity int
	maxDepth int
	depth    int
	entities []SpatialEntity
	divided  bool
	children [4]*QuadTree
}

func NewQuadTree(bounds SpatialBounds, capacity, maxDepth int) *QuadTree {
	return &QuadTree{
		bounds:   bounds,
		capacity: capacity,
		maxDepth: maxDepth,
	}
}

func (qt *QuadTree) Insert(entity SpatialEntity) bool {
	if !qt.contains(entity) {
		return false
	}
	if len(qt.entities) < qt.capacity || qt.depth >= qt.maxDepth {
		qt.entities = append(qt.entities, entity)
		return true
	}
	if !qt.divided {
		qt.subdivide()
	}
	for i := 0; i < 4; i++ {
		if qt.children[i].Insert(entity) {
			return true
		}
	}
	return false
}

func (qt *QuadTree) Remove(entity SpatialEntity) bool {
	for i, e := range qt.entities {
		if e.GetID() == entity.GetID() {
			qt.entities = append(qt.entities[:i], qt.entities[i+1:]...)
			return true
		}
	}
	if qt.divided {
		for i := 0; i < 4; i++ {
			if qt.children[i].Remove(entity) {
				return true
			}
		}
	}
	return false
}

func (qt *QuadTree) QueryRange(bounds SpatialBounds) []SpatialEntity {
	var result []SpatialEntity
	if !qt.intersects(bounds) {
		return result
	}
	for _, e := range qt.entities {
		if e.GetX() >= bounds.X && e.GetX() <= bounds.X+bounds.W &&
			e.GetY() >= bounds.Y && e.GetY() <= bounds.Y+bounds.H {
			result = append(result, e)
		}
	}
	if qt.divided {
		for i := 0; i < 4; i++ {
			result = append(result, qt.children[i].QueryRange(bounds)...)
		}
	}
	return result
}

func (qt *QuadTree) Clear() {
	qt.entities = nil
	qt.divided = false
	qt.children = [4]*QuadTree{}
}

func (qt *QuadTree) contains(entity SpatialEntity) bool {
	return entity.GetX() >= qt.bounds.X && entity.GetX() <= qt.bounds.X+qt.bounds.W &&
		entity.GetY() >= qt.bounds.Y && entity.GetY() <= qt.bounds.Y+qt.bounds.H
}

func (qt *QuadTree) intersects(bounds SpatialBounds) bool {
	return !(bounds.X > qt.bounds.X+qt.bounds.W ||
		bounds.X+bounds.W < qt.bounds.X ||
		bounds.Y > qt.bounds.Y+qt.bounds.H ||
		bounds.Y+bounds.H < qt.bounds.Y)
}

func (qt *QuadTree) subdivide() {
	halfW := qt.bounds.W / 2
	halfH := qt.bounds.H / 2
	qt.children[0] = NewQuadTree(SpatialBounds{qt.bounds.X, qt.bounds.Y, halfW, halfH}, qt.capacity, qt.maxDepth)
	qt.children[1] = NewQuadTree(SpatialBounds{qt.bounds.X + halfW, qt.bounds.Y, halfW, halfH}, qt.capacity, qt.maxDepth)
	qt.children[2] = NewQuadTree(SpatialBounds{qt.bounds.X, qt.bounds.Y + halfH, halfW, halfH}, qt.capacity, qt.maxDepth)
	qt.children[3] = NewQuadTree(SpatialBounds{qt.bounds.X + halfW, qt.bounds.Y + halfH, halfW, halfH}, qt.capacity, qt.maxDepth)
	qt.divided = true

	for _, e := range qt.entities {
		for i := 0; i < 4; i++ {
			if qt.children[i].Insert(e) {
				break
			}
		}
	}
	qt.entities = nil
}
