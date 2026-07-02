package core

type LifecycleState int

const (
	LifecycleUnallocated      LifecycleState = 0
	LifecycleAllocated        LifecycleState = 1
	LifecycleInitializing     LifecycleState = 2
	LifecycleActive           LifecycleState = 3
	LifecycleSuspended        LifecycleState = 4
	LifecycleMarkedForRemoval LifecycleState = 5
	LifecycleDestroyed        LifecycleState = 6
	LifecycleReturnedToPool   LifecycleState = 7
)
