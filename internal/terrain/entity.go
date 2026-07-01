package terrain

import rl "github.com/gen2brain/raylib-go/raylib"

type EntityFlags uint32

const (
	FlagNone        EntityFlags = 0
	FlagPowered     EntityFlags = 1 << 0
	FlagWatered     EntityFlags = 1 << 1
	FlagConnected   EntityFlags = 1 << 2
	FlagOccupied    EntityFlags = 1 << 3
	FlagAbandoned   EntityFlags = 1 << 4
	FlagConstructed EntityFlags = 1 << 5
	FlagHasRoad     EntityFlags = 1 << 6
	FlagParked      EntityFlags = 1 << 7
	FlagRemoved     EntityFlags = 1 << 8
)

type OwnerType uint16

const (
	OwnerNone        OwnerType = 0
	OwnerBuilding    OwnerType = 1
	OwnerVehicle     OwnerType = 2
	OwnerTree        OwnerType = 3
	OwnerTransport   OwnerType = 4
	OwnerCitizen     OwnerType = 5
)

type Entity struct {
	ID           uint32
	Position     rl.Vector3
	Rotation     rl.Vector4
	Bounds       rl.BoundingBox
	Flags        EntityFlags
	Owner        OwnerType
	LODLevel     uint16
	Lifecycle    LifecycleState
	CreatedAt    int32
	RemovalTimer int32
	Dirty        bool
}

func NewEntity(id uint32, x, y, z float32, owner OwnerType) Entity {
	return Entity{
		ID:       id,
		Position: rl.NewVector3(x, y, z),
		Rotation: rl.NewVector4(0, 1, 0, 0),
		Owner:    owner,
	}
}

func (e *Entity) SetPosition(x, y, z float32) {
	e.Position.X = x
	e.Position.Y = y
	e.Position.Z = z
}

func (e *Entity) SetRotationY(degrees float32) {
	e.Rotation = rl.NewVector4(0, 1, 0, degrees*rl.Deg2rad)
}

func (e *Entity) HasFlag(flag EntityFlags) bool {
	return e.Flags&flag != 0
}

func (e *Entity) SetFlag(flag EntityFlags) {
	e.Flags |= flag
}

func (e *Entity) ClearFlag(flag EntityFlags) {
	e.Flags &^= flag
}

func (e *Entity) UpdateBounds(width, depth, height float32) {
	halfW := width / 2
	halfD := depth / 2
	e.Bounds = rl.BoundingBox{
		Min: rl.NewVector3(e.Position.X-halfW, e.Position.Y, e.Position.Z-halfD),
		Max: rl.NewVector3(e.Position.X+halfW, e.Position.Y+height, e.Position.Z+halfD),
	}
}
