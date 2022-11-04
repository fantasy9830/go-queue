package queue

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/fantasy9830/go-graceful"
	"github.com/fantasy9830/go-queue/dispatcher"
	"github.com/fantasy9830/go-retry"
)

type Queue struct {
	inShutdown atomic.Bool
	once       sync.Once
	ctx        context.Context
	cancel     context.CancelFunc
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
		ctx:        opt.ctx,
		cancel:     opt.cancel,
		jobChan:    make(chan Job, opt.queueSize),
		dispatcher: dispatcher.NewDispatcher(dispatcher.WithMaxWorkers(opt.maxWorkers)),
		graceful:   graceful.NewManager(graceful.WithContext(opt.ctx)),
	}

	q.graceful.Go(q.Start)
	q.graceful.RegisterOnShutdown(q.OnShutdown)
	q.graceful.RegisterOnShutdown(q.dispatcher.OnShutdown)

	go func() {
		<-q.graceful.Done()
		opt.cancel()
	}()

	return q
}

func (q *Queue) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			q.dispatcher.Dispatch()

			<-q.dispatcher.WaitReady()

			job, ok := <-q.jobChan
			if !ok {
				return
			}

			q.dispatcher.IncWorker()
			q.graceful.Go(func(ctx context.Context) {
				defer func() {
					q.dispatcher.DecWorker()
					q.dispatcher.Dispatch()
				}()

				opt := []retry.OptionFunc{
					retry.WithContext(ctx),
					retry.MaxRetries(job.maxRetries),
					retry.WithBackoff(job.backoffFunc),
				}

				if err := retry.Do(job.TaskFunc, opt...); err != nil {
					log.Print(err.Error())
				}
			})
		}
	}
}

func (q *Queue) Done() <-chan struct{} {
	return q.ctx.Done()
}

func (q *Queue) AddJob(taskFunc func(context.Context) error, optFuncs ...func(*Job)) error {
	if q.shuttingDown() {
		return errors.New("queue has been closed and released")
	}

	job := NewJob(taskFunc, optFuncs...)

	select {
	case q.jobChan <- job:
	default:
	}

	return nil
}

func (q *Queue) OnShutdown() {
	q.once.Do(func() {
		q.setShuttingDown()
		close(q.jobChan)
	})
}

func (q *Queue) shuttingDown() bool {
	return q.inShutdown.Load()
}

func (q *Queue) setShuttingDown() {
	q.inShutdown.Store(true)
}
