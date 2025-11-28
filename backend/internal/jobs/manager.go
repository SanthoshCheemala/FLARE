package jobs

import (
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusRunning   Status = "RUNNING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
	StatusCancelled Status = "CANCELLED"
)

type Phase string

const (
	PhaseServerInit    Phase = "server_init"
	PhaseClientEncrypt Phase = "client_encrypt"
	PhaseIntersection  Phase = "intersection"
	PhasePersist       Phase = "persist"
	PhaseComplete      Phase = "complete"
)

type Progress struct {
	Phase     Phase             `json:"phase"`
	Percent   int               `json:"percent"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Metrics   map[string]string `json:"metrics,omitempty"`
}

type ScreeningJob struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Status            Status     `json:"status"`
	Progress          []Progress `json:"progress"`
	CustomerListID    int64      `json:"customerListId"`
	SanctionListIDs   []int64    `json:"sanctionListIds"`
	ResultIDs         []int64    `json:"resultIds,omitempty"`
	MatchCount        int        `json:"matchCount"`
	CustomerCount     int        `json:"customerCount"`
	SanctionCount     int        `json:"sanctionCount"`
	StartedAt         time.Time  `json:"startedAt,omitempty"`
	FinishedAt        time.Time  `json:"finishedAt,omitempty"`
	Error             string     `json:"error,omitempty"`
	CreatedBy         int64      `json:"createdBy"`
	WorkerCount       int        `json:"workerCount"`
	MemoryEstimateMB  float64    `json:"memoryEstimateMb"`
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	progressListeners []chan Progress
}

type Manager struct {
	mu            sync.RWMutex
	jobs          map[string]*ScreeningJob
	maxConcurrent int
	running       int
}

func NewManager(maxConcurrent int) *Manager {
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}
	return &Manager{
		jobs:          make(map[string]*ScreeningJob),
		maxConcurrent: maxConcurrent,
	}
}

func (m *Manager) Create(id, name string, customerListID int64, sanctionListIDs []int64, createdBy int64) *ScreeningJob {
	ctx, cancel := context.WithCancel(context.Background())
	job := &ScreeningJob{
		ID:                id,
		Name:              name,
		Status:            StatusPending,
		Progress:          []Progress{},
		CustomerListID:    customerListID,
		SanctionListIDs:   sanctionListIDs,
		CreatedBy:         createdBy,
		ctx:               ctx,
		cancel:            cancel,
		progressListeners: []chan Progress{},
	}

	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	return job
}

func (m *Manager) Get(id string) *ScreeningJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobs[id]
}

func (m *Manager) List() []*ScreeningJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]*ScreeningJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (m *Manager) CanStart() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running < m.maxConcurrent
}

func (m *Manager) IncrementRunning() {
	m.mu.Lock()
	m.running++
	m.mu.Unlock()
}

func (m *Manager) DecrementRunning() {
	m.mu.Lock()
	if m.running > 0 {
		m.running--
	}
	m.mu.Unlock()
}

func (j *ScreeningJob) AddProgress(phase Phase, percent int, message string, metrics map[string]string) {
	j.mu.Lock()
	p := Progress{
		Phase:     phase,
		Percent:   percent,
		Message:   message,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}
	j.Progress = append(j.Progress, p)

	// Notify listeners
	for _, listener := range j.progressListeners {
		select {
		case listener <- p:
		default:
			// Don't block if listener is slow
		}
	}
	j.mu.Unlock()
}

func (j *ScreeningJob) Subscribe() chan Progress {
	j.mu.Lock()
	defer j.mu.Unlock()

	ch := make(chan Progress, 10)
	j.progressListeners = append(j.progressListeners, ch)
	return ch
}

func (j *ScreeningJob) Unsubscribe(ch chan Progress) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i, listener := range j.progressListeners {
		if listener == ch {
			j.progressListeners = append(j.progressListeners[:i], j.progressListeners[i+1:]...)
			close(ch)
			break
		}
	}
}

func (j *ScreeningJob) SetStatus(status Status) {
	j.mu.Lock()
	j.Status = status
	if status == StatusRunning && j.StartedAt.IsZero() {
		j.StartedAt = time.Now()
	}
	if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
		j.FinishedAt = time.Now()

		// Close all listeners
		for _, listener := range j.progressListeners {
			close(listener)
		}
		j.progressListeners = nil
	}
	j.mu.Unlock()
}

func (j *ScreeningJob) SetError(err error) {
	j.mu.Lock()
	if err != nil {
		j.Error = err.Error()
	}
	j.mu.Unlock()
}

func (j *ScreeningJob) SetResults(resultIDs []int64, matchCount int) {
	j.mu.Lock()
	j.ResultIDs = resultIDs
	j.MatchCount = matchCount
	j.mu.Unlock()
}

func (j *ScreeningJob) SetCounts(customerCount, sanctionCount int) {
	j.mu.Lock()
	j.CustomerCount = customerCount
	j.SanctionCount = sanctionCount
	j.mu.Unlock()
}

func (j *ScreeningJob) SetWorkerInfo(workerCount int, memoryMB float64) {
	j.mu.Lock()
	j.WorkerCount = workerCount
	j.MemoryEstimateMB = memoryMB
	j.mu.Unlock()
}

func (j *ScreeningJob) Cancel() {
	j.cancel()
	j.SetStatus(StatusCancelled)
}

func (j *ScreeningJob) Context() context.Context {
	return j.ctx
}

func (j *ScreeningJob) GetSnapshot() ScreeningJob {
	j.mu.RLock()
	defer j.mu.RUnlock()

	// Create a copy without the internal fields
	return ScreeningJob{
		ID:               j.ID,
		Name:             j.Name,
		Status:           j.Status,
		Progress:         append([]Progress{}, j.Progress...),
		CustomerListID:   j.CustomerListID,
		SanctionListIDs:  append([]int64{}, j.SanctionListIDs...),
		ResultIDs:        append([]int64{}, j.ResultIDs...),
		MatchCount:       j.MatchCount,
		CustomerCount:    j.CustomerCount,
		SanctionCount:    j.SanctionCount,
		StartedAt:        j.StartedAt,
		FinishedAt:       j.FinishedAt,
		Error:            j.Error,
		CreatedBy:        j.CreatedBy,
		WorkerCount:      j.WorkerCount,
		MemoryEstimateMB: j.MemoryEstimateMB,
	}
}
