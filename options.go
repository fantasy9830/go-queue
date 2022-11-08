package queue

import (
	"context"
	"runtime"
)

type OptionFunc func(*Options)

type Options struct {
	ctx        context.Context
	queueSize  uint64
	maxWorkers uint64
}

func DefaultOption() *Options {
	return &Options{
		ctx:        context.Background(),
		queueSize:  4096,
		maxWorkers: uint64(runtime.NumCPU()),
	}
}

func WithContext(parent context.Context) OptionFunc {
	return func(m *Options) {
		m.ctx = parent
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
