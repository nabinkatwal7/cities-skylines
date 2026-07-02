package core

import "time"

type UpdateGroup int

const (
	GroupFast     UpdateGroup = 0
	GroupMedium   UpdateGroup = 1
	GroupSlow     UpdateGroup = 2
	GroupVerySlow UpdateGroup = 3
)

type SchedPriority int

const (
	SchedPriorityCritical SchedPriority = 0
	SchedPriorityHigh     SchedPriority = 1
	SchedPriorityMedium   SchedPriority = 2
	SchedPriorityLow      SchedPriority = 3
)

type UpdateTask struct {
	Name     string
	Callback func(dt float64)
	Priority SchedPriority
	BudgetMs float64
}

type Scheduler struct {
	groups    [4][]UpdateTask
	budgets   [4]float64
	intervals [4]int32
	counters  [4]int32

	frameCount int32
	deferred   []UpdateTask
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		budgets:   [4]float64{3, 5, 10, 20},
		intervals: [4]int32{1, 4, 30, 120},
	}
}

func (s *Scheduler) Register(group UpdateGroup, task UpdateTask) {
	s.groups[group] = append(s.groups[group], task)
}

func (s *Scheduler) Unregister(group UpdateGroup, name string) {
	idx := -1
	for i, t := range s.groups[group] {
		if t.Name == name {
			idx = i
			break
		}
	}
	if idx >= 0 {
		s.groups[group] = append(s.groups[group][:idx], s.groups[group][idx+1:]...)
	}
}

func (s *Scheduler) SetInterval(group UpdateGroup, interval int32) {
	s.intervals[group] = interval
}

func (s *Scheduler) SetBudget(group UpdateGroup, ms float64) {
	s.budgets[group] = ms
}

func (s *Scheduler) RunGroup(group UpdateGroup, dt float64) {
	if s.counters[group] < s.intervals[group]-1 {
		s.counters[group]++
		return
	}
	s.counters[group] = 0

	tasks := s.groups[group]
	budgetMs := s.budgets[group]
	used := 0.0

	for _, task := range tasks {
		if task.BudgetMs > 0 && used+task.BudgetMs > budgetMs && task.Priority >= SchedPriorityMedium {
			s.deferred = append(s.deferred, task)
			continue
		}
		taskStart := time.Now()
		task.Callback(dt)
		elapsed := float64(time.Since(taskStart).Microseconds()) / 1000.0
		used += elapsed
		if used > budgetMs && task.Priority >= SchedPriorityMedium {
			break
		}
	}
}

func (s *Scheduler) RunDeferred(dt float64) {
	if len(s.deferred) == 0 {
		return
	}
	budget := s.budgets[GroupSlow]
	start := time.Now()
	var remaining []UpdateTask
	for _, task := range s.deferred {
		if float64(time.Since(start).Microseconds())/1000.0 > budget {
			remaining = append(remaining, task)
			continue
		}
		task.Callback(dt)
	}
	s.deferred = remaining
}

func (s *Scheduler) RunAll(dt float64) {
	s.frameCount++
	s.RunGroup(GroupFast, dt)
	s.RunGroup(GroupMedium, dt)
	s.RunGroup(GroupSlow, dt)
	s.RunGroup(GroupVerySlow, dt)
	s.RunDeferred(dt)
}

func (s *Scheduler) FrameCount() int32 {
	return s.frameCount
}

func (s *Scheduler) Clear() {
	for i := range s.groups {
		s.groups[i] = nil
	}
	s.deferred = nil
}
