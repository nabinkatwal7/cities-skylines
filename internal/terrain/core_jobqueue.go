package terrain

import "time"

type JobState int

const (
	JobCreated   JobState = 0
	JobScheduled JobState = 1
	JobExecuted  JobState = 2
	JobValidated JobState = 3
	JobCommitted JobState = 4
	JobFailed    JobState = 5
)

type Job struct {
	ID           int
	Name         string
	State        JobState
	Execute      func()
	Validate     func() bool
	Commit       func()
	Dependencies []int
}

type JobQueue struct {
	jobs       map[int]*Job
	jobCounter int
	pending    []*Job
	ready      []*Job
	budgetMs   float64
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs:     make(map[int]*Job),
		budgetMs: 2,
	}
}

func (jq *JobQueue) SetBudget(ms float64) { jq.budgetMs = ms }

func (jq *JobQueue) Create(name string, execute func(), deps []int) int {
	id := jq.jobCounter
	jq.jobCounter++
	job := &Job{
		ID:           id,
		Name:         name,
		State:        JobCreated,
		Execute:      execute,
		Dependencies: deps,
	}
	jq.jobs[id] = job
	jq.pending = append(jq.pending, job)
	return id
}

func (jq *JobQueue) Process() int {
	start := time.Now()
	processed := 0
	jq.scheduleReady()

	for len(jq.ready) > 0 {
		if time.Since(start).Seconds()*1000 >= jq.budgetMs {
			break
		}
		job := jq.ready[0]
		jq.ready = jq.ready[1:]

		job.Execute()
		job.State = JobExecuted
		processed++

		if job.Validate != nil && !job.Validate() {
			job.State = JobFailed
			continue
		}
		job.State = JobValidated

		if job.Commit != nil {
			job.Commit()
		}
		job.State = JobCommitted
	}

	return processed
}

func (jq *JobQueue) scheduleReady() {
	var remaining []*Job
	for _, job := range jq.pending {
		depsMet := true
		for _, depID := range job.Dependencies {
			if dep, ok := jq.jobs[depID]; !ok || dep.State < JobCommitted {
				depsMet = false
				break
			}
		}
		if depsMet {
			job.State = JobScheduled
			jq.ready = append(jq.ready, job)
		} else {
			remaining = append(remaining, job)
		}
	}
	jq.pending = remaining
}

func (jq *JobQueue) PendingCount() int {
	return len(jq.pending) + len(jq.ready)
}

func (jq *JobQueue) Clear() {
	for k := range jq.jobs {
		delete(jq.jobs, k)
	}
	jq.pending = nil
	jq.ready = nil
}
