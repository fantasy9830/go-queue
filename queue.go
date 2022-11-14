package queue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/fantasy9830/go-graceful"
	"github.com/fantasy9830/go-queue/dispatcher"
	"github.com/fantasy9830/go-retry"
)

type Queue struct {
	inShutdown atomic.Bool
	once       sync.Once
	jobChan    chan Job
	dispatcher dispatcher.IDispatcher
	graceful   graceful.GracefulManager
	quit       chan struct{}
}

func NewQueue(optFuncs ...OptionFunc) *Queue {
	opt := DefaultOption()
	for _, applyFunc := range optFuncs {
		applyFunc(opt)
	}

	q := &Queue{
		jobChan:    make(chan Job, opt.queueSize),
		quit:       make(chan struct{}, 1),
		dispatcher: dispatcher.NewDispatcher(dispatcher.WithMaxWorkers(opt.maxWorkers)),
		graceful:   graceful.NewManager(graceful.WithContext(opt.ctx)),
	}

	q.graceful.Go(q.start)
	q.graceful.RegisterOnShutdown(q.onShutdown)
	q.graceful.RegisterOnShutdown(q.dispatcher.OnShutdown)

	return q
}

func (q *Queue) Done() <-chan struct{} {
	return q.graceful.Done()
}

func (q *Queue) Errors() <-chan error {
	return q.graceful.Errors()
}

func (q *Queue) AddJob(taskFunc func(context.Context) error, optFuncs ...func(*Job)) error {
	if q.shuttingDown() {
		return errors.New("queue has been closed and released")
	}

	job := NewJob(taskFunc, optFuncs...)

	select {
	case <-q.quit:
		return nil
	case q.jobChan <- job:
	}

	return nil
}

func (q *Queue) start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			q.dispatcher.Dispatch()

			select {
			case <-q.quit:
				return nil
			case _, ok := <-q.dispatcher.WaitReady():
				if !ok {
					return nil
				}
			}

			select {
			case <-q.quit:
				return nil
			case job, ok := <-q.jobChan:
				if !ok {
					return nil
				}

				q.dispatcher.IncWorker()
				q.graceful.Go(func(ctx context.Context) error {
					defer func() {
						q.dispatcher.DecWorker()
						q.dispatcher.Dispatch()
					}()

					opt := []retry.OptionFunc{
						retry.WithContext(ctx),
						retry.MaxRetries(job.maxRetries),
						retry.WithBackoff(job.backoffFunc),
					}

					return retry.Do(job.TaskFunc, opt...)
				})
			}
		}
	}
}

func (q *Queue) onShutdown() error {
	q.once.Do(func() {
		q.setShuttingDown()
		q.quit <- struct{}{}
		close(q.quit)
	})

	return nil
}

func (q *Queue) shuttingDown() bool {
	return q.inShutdown.Load()
}

func (q *Queue) setShuttingDown() {
	q.inShutdown.Store(true)
}
