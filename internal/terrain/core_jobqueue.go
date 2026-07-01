package terrain

import (
	"runtime"
	"sync"
	"time"
)

type JobState int

const (
	JobCreated    JobState = 0
	JobScheduled  JobState = 1
	JobExecuting  JobState = 2
	JobExecuted   JobState = 3
	JobValidated  JobState = 4
	JobCommitted  JobState = 5
	JobFailed     JobState = 6
)

type JobResult struct {
	Data  any
	Error error
}

type Job struct {
	ID           int
	Name         string
	State        JobState
	Execute      func() JobResult
	Validate     func(JobResult) bool
	Commit       func(JobResult)
	Dependencies []int
	result       JobResult
}

type JobQueue struct {
	mu         sync.Mutex
	jobs       map[int]*Job
	jobCounter int
	pending    []*Job
	ready      []*Job
	inFlight   int
	budgetMs   float64

	workerWg  sync.WaitGroup
	workChan  chan *Job
	resultBuf []*Job
	done      chan struct{}
	workers   int
}

func NewJobQueue() *JobQueue {
	n := runtime.NumCPU()
	if n < 1 {
		n = 1
	}
	jq := &JobQueue{
		jobs:     make(map[int]*Job),
		budgetMs: 4,
		workers:  n - 1,
		workChan: make(chan *Job, 256),
		done:     make(chan struct{}),
	}
	if jq.workers < 1 {
		jq.workers = 1
	}
	jq.startWorkers()
	return jq
}

func (jq *JobQueue) startWorkers() {
	for i := 0; i < jq.workers; i++ {
		jq.workerWg.Add(1)
		go func() {
			defer jq.workerWg.Done()
			for job := range jq.workChan {
				job.State = JobExecuting
				job.result = job.Execute()
				job.State = JobExecuted
				jq.mu.Lock()
				jq.resultBuf = append(jq.resultBuf, job)
				jq.mu.Unlock()
			}
		}()
	}
}

func (jq *JobQueue) Stop() {
	close(jq.workChan)
	jq.workerWg.Wait()
}

func (jq *JobQueue) SetBudget(ms float64) { jq.budgetMs = ms }

func (jq *JobQueue) Create(name string, execute func() JobResult, deps []int) int {
	jq.mu.Lock()
	defer jq.mu.Unlock()
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

func (jq *JobQueue) CreateFull(name string, execute func() JobResult, validate func(JobResult) bool, commit func(JobResult), deps []int) int {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	id := jq.jobCounter
	jq.jobCounter++
	job := &Job{
		ID:           id,
		Name:         name,
		State:        JobCreated,
		Execute:      execute,
		Validate:     validate,
		Commit:       commit,
		Dependencies: deps,
	}
	jq.jobs[id] = job
	jq.pending = append(jq.pending, job)
	return id
}

func (jq *JobQueue) Process() int {
	start := time.Now()
	processed := 0

	for {
		elapsed := func() float64 { return float64(time.Since(start).Microseconds()) / 1000.0 }()

		jq.mu.Lock()
		var batch []*Job
		n := len(jq.resultBuf)
		if n > 0 {
			batch = make([]*Job, n)
			copy(batch, jq.resultBuf)
			jq.resultBuf = jq.resultBuf[:0]
		}

		var pendingLeft int
		for _, job := range jq.pending {
			depsMet := true
			for _, depID := range job.Dependencies {
				if dep, ok := jq.jobs[depID]; !ok || dep.State < JobCommitted {
					depsMet = false
					break
				}
			}
			if depsMet && elapsed < jq.budgetMs {
				job.State = JobScheduled
				jq.ready = append(jq.ready, job)
			} else {
				pendingLeft++
			}
		}
		reordered := make([]*Job, 0, len(jq.pending))
		for _, job := range jq.pending {
			if job.State == JobCreated {
				reordered = append(reordered, job)
			}
		}
		jq.pending = reordered

		for _, job := range jq.ready {
			select {
			case jq.workChan <- job:
			default:
			}
		}
		jq.ready = nil
		jq.mu.Unlock()

		for _, job := range batch {
			if elapsed >= jq.budgetMs {
				jq.mu.Lock()
				rest := make([]*Job, len(batch)-processed)
				copy(rest, batch[processed:])
				jq.resultBuf = append(jq.resultBuf, rest...)
				jq.mu.Unlock()
				return processed
			}

			if job.Validate != nil && !job.Validate(job.result) {
				job.State = JobFailed
				processed++
				continue
			}
			job.State = JobValidated

			if job.Commit != nil {
				job.Commit(job.result)
			}
			job.State = JobCommitted
			processed++
		}

		if elapsed >= jq.budgetMs {
			break
		}

		jq.mu.Lock()
		rlen := len(jq.resultBuf)
		jq.mu.Unlock()
		if rlen == 0 && pendingLeft == 0 {
			break
		}
	}

	return processed
}

func (jq *JobQueue) scheduleReady() {
	jq.mu.Lock()
	defer jq.mu.Unlock()

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

	for _, job := range jq.ready {
		select {
		case jq.workChan <- job:
			jq.inFlight++
		default:
			break
		}
	}
	jq.ready = nil
}

func (jq *JobQueue) PendingCount() int {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	return len(jq.pending) + len(jq.ready) + jq.inFlight + len(jq.resultBuf)
}

func (jq *JobQueue) Clear() {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	for k := range jq.jobs {
		delete(jq.jobs, k)
	}
	jq.pending = nil
	jq.ready = nil
	jq.resultBuf = nil
	jq.inFlight = 0
}
