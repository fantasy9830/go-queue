package queue

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/fantasy9830/go-graceful"
	"github.com/fantasy9830/go-queue/dispatcher"
	"github.com/fantasy9830/go-retry"
)

type Queue struct {
	inShutdown atomic.Bool
	jobChan    chan Job
	dispatcher dispatcher.IDispatcher
	graceful   graceful.GracefulManager
}

func NewQueue(optFuncs ...OptionFunc) *Queue {
	opt := DefaultOption()
	for _, applyFunc := range optFuncs {
		applyFunc(opt)
	}

	q := &Queue{
		jobChan:    make(chan Job, opt.queueSize),
		dispatcher: dispatcher.NewDispatcher(dispatcher.WithMaxWorkers(opt.maxWorkers)),
		graceful:   graceful.NewManager(graceful.WithContext(opt.ctx)),
	}

	q.graceful.Go(q.start)
	q.graceful.RegisterOnShutdown(q.onShutdown)

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

	q.jobChan <- job

	return nil
}

func (q *Queue) start(ctx context.Context) error {
	for {
		q.dispatcher.Dispatch()

		<-q.dispatcher.WaitReady()

		select {
		case <-ctx.Done():
			return nil
		case job := <-q.jobChan:
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

func (q *Queue) onShutdown() error {
	q.setShuttingDown()
	return q.dispatcher.OnShutdown()
}

func (q *Queue) shuttingDown() bool {
	return q.inShutdown.Load()
}

func (q *Queue) setShuttingDown() {
	q.inShutdown.Store(true)
}
