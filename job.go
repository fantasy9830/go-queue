package queue

import (
	"context"
	"time"
)

type BackoffFunc func(uint) time.Duration

type Job struct {
	TaskFunc    func(context.Context) error
	maxRetries  uint
	backoffFunc func(uint) time.Duration
}

func NewJob(taskFunc func(context.Context) error, optFuncs ...func(*Job)) Job {
	defaultJob := &Job{
		TaskFunc:    taskFunc,
		maxRetries:  10,
		backoffFunc: BackoffLinear(3 * time.Second),
	}

	for _, applyFunc := range optFuncs {
		applyFunc(defaultJob)
	}

	return *defaultJob
}

func MaxRetries(maxRetries uint) func(*Job) {
	return func(job *Job) {
		job.maxRetries = maxRetries
	}
}

func WithBackoff(backoffFunc BackoffFunc) func(*Job) {
	return func(job *Job) {
		job.backoffFunc = backoffFunc
	}
}

func BackoffLinear(duration time.Duration) BackoffFunc {
	return func(attempt uint) time.Duration {
		return duration
	}
}

func BackoffExponential(duration time.Duration) BackoffFunc {
	return func(attempt uint) time.Duration {
		return ((1 << attempt) >> 1) * duration
	}
}
