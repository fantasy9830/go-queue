package dispatcher

import (
	"sync"
	"sync/atomic"
)

type IDispatcher interface {
	IncWorker()
	DecWorker()
	GetWorkerCount() uint64
	UpdateMaxWorkers(count uint64)
	GetMaxWorkers() uint64
	Dispatch()
	WaitReady() <-chan struct{}
}

type Dispatcher struct {
	mu          sync.Mutex
	maxWorkers  atomic.Uint64
	workerCount atomic.Uint64
	ready       chan struct{}
}

func NewDispatcher(opts ...func(*Dispatcher)) IDispatcher {
	d := &Dispatcher{
		ready: make(chan struct{}, 1),
	}

	for _, f := range opts {
		f(d)
	}

	return d
}

func WithMaxWorkers(maxCount uint64) func(*Dispatcher) {
	return func(d *Dispatcher) {
		d.maxWorkers.Store(maxCount)
	}
}

func (d *Dispatcher) IncWorker() {
	d.workerCount.Add(1)
}

func (d *Dispatcher) DecWorker() {
	d.workerCount.Add(^uint64(0))
}

func (d *Dispatcher) GetWorkerCount() uint64 {
	return d.workerCount.Load()
}

func (d *Dispatcher) UpdateMaxWorkers(count uint64) {
	d.maxWorkers.Store(count)
	d.Dispatch()
}

func (d *Dispatcher) GetMaxWorkers() uint64 {
	return d.maxWorkers.Load()
}

func (d *Dispatcher) Dispatch() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.GetWorkerCount() >= d.GetMaxWorkers() {
		return
	}

	select {
	case d.ready <- struct{}{}:
	default:
	}
}

func (d *Dispatcher) WaitReady() <-chan struct{} {
	return d.ready
}
