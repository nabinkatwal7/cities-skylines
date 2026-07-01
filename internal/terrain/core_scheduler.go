package terrain

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
}

type Scheduler struct {
	groups  [][]UpdateTask
	budgets []float64
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		groups:  make([][]UpdateTask, 4),
		budgets: []float64{3, 5, 10, 20},
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

func (s *Scheduler) SetBudget(group UpdateGroup, ms float64) {
	s.budgets[group] = ms
}

func (s *Scheduler) RunGroup(group UpdateGroup, dt float64) {
	tasks := s.groups[group]
	if len(tasks) == 0 {
		return
	}

	sortTasks(tasks)

	budgetMs := s.budgets[group]
	start := time.Now()

	for _, task := range tasks {
		task.Callback(dt)
		if time.Since(start).Milliseconds() > int64(budgetMs) && task.Priority >= SchedPriorityLow {
			break
		}
	}
}

func (s *Scheduler) Clear() {
	for i := range s.groups {
		s.groups[i] = nil
	}
}

func sortTasks(tasks []UpdateTask) {
	for i := 1; i < len(tasks); i++ {
		key := tasks[i]
		j := i - 1
		for j >= 0 && tasks[j].Priority > key.Priority {
			tasks[j+1] = tasks[j]
			j--
		}
		tasks[j+1] = key
	}
}
