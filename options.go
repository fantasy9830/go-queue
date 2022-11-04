package queue

import (
	"context"
	"runtime"
)

type OptionFunc func(*Options)

type Options struct {
	ctx        context.Context
	cancel     context.CancelFunc
	queueSize  uint64
	maxWorkers uint64
}

func DefaultOption() *Options {
	ctx, cancel := context.WithCancel(context.Background())

	return &Options{
		ctx:        ctx,
		cancel:     cancel,
		queueSize:  4096,
		maxWorkers: uint64(runtime.NumCPU()),
	}
}

func WithContext(parent context.Context) OptionFunc {
	return func(m *Options) {
		ctx, cancel := context.WithCancel(parent)

		m.ctx = ctx
		m.cancel = cancel
	}
}

func WithQueueSize(size uint64) OptionFunc {
	return func(opt *Options) {
		opt.queueSize = size
	}
}

func WithMaxWorkers(maxCount uint64) OptionFunc {
	return func(opt *Options) {
		opt.maxWorkers = maxCount
	}
}
