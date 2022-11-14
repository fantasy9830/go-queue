package queue_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fantasy9830/go-queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suit struct {
	suite.Suite
}

func TestQueue(t *testing.T) {
	suite.Run(t, new(Suit))
}

func (s *Suit) TestNewQueue() {
	var count atomic.Int32

	parent, cancel := context.WithCancel(context.Background())

	opts := []queue.OptionFunc{
		queue.WithContext(parent),
		queue.WithMaxWorkers(4),
		queue.WithQueueSize(4),
	}
	q := queue.NewQueue(opts...)

	go func() {
		for err := range q.Errors() {
			assert.Equal(s.T(), "test", err.Error())
		}
	}()

	err := q.AddJob(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		count.Add(1)
		return nil
	})
	assert.NoError(s.T(), err)

	err = q.AddJob(func(ctx context.Context) error {
		count.Add(1)
		return errors.New("test")
	}, queue.MaxRetries(1), queue.WithBackoff(queue.BackoffLinear(0*time.Second)))
	assert.NoError(s.T(), err)

	err = q.AddJob(func(ctx context.Context) error {
		count.Add(1)
		return errors.New("test")
	}, queue.MaxRetries(2), queue.WithBackoff(queue.BackoffExponential(1*time.Second)))
	assert.NoError(s.T(), err)

	err = q.AddJob(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		count.Add(1)
		return errors.New("test")
	})
	assert.NoError(s.T(), err)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-q.Done()

	assert.Equal(s.T(), int32(4), count.Load())
}

func (s *Suit) TestInShuttingDown() {
	var count atomic.Int32

	parent, cancel := context.WithCancel(context.Background())

	opts := []queue.OptionFunc{
		queue.WithContext(parent),
	}
	q := queue.NewQueue(opts...)
	err := q.AddJob(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		count.Add(1)
		return nil
	})
	assert.NoError(s.T(), err)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := q.AddJob(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			count.Add(1)
			return nil
		})
		assert.Equal(s.T(), "queue has been closed and released", err.Error())

		err = q.AddJob(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			count.Add(1)
			return nil
		})
		assert.Equal(s.T(), "queue has been closed and released", err.Error())
	}()

	<-q.Done()

	assert.Equal(s.T(), int32(1), count.Load())
}

func (s *Suit) TestAfterShuttingDown() {
	var count atomic.Int32

	parent, cancel := context.WithCancel(context.Background())

	opts := []queue.OptionFunc{
		queue.WithContext(parent),
	}
	q := queue.NewQueue(opts...)
	cancel()

	time.Sleep(100 * time.Millisecond)

	err := q.AddJob(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		count.Add(1)
		return nil
	})
	assert.Equal(s.T(), "queue has been closed and released", err.Error())

	<-q.Done()

	assert.Equal(s.T(), int32(0), count.Load())
}
