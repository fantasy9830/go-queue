package dispatcher_test

import (
	"testing"

	"github.com/fantasy9830/go-queue/dispatcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suit struct {
	suite.Suite
}

func TestDispatcher(t *testing.T) {
	suite.Run(t, new(Suit))
}

func (s *Suit) TestWithMaxWorkers() {
	d := dispatcher.NewDispatcher(dispatcher.WithMaxWorkers(100))

	assert.Equal(s.T(), uint64(100), d.GetMaxWorkers())
}

func (s *Suit) TestWorker() {
	d := dispatcher.NewDispatcher()
	assert.Equal(s.T(), uint64(0), d.GetWorkerCount())

	d.IncWorker()
	assert.Equal(s.T(), uint64(1), d.GetWorkerCount())

	d.DecWorker()
	assert.Equal(s.T(), uint64(0), d.GetWorkerCount())

	d.IncWorker()
	assert.Equal(s.T(), uint64(1), d.GetWorkerCount())

	d.IncWorker()
	assert.Equal(s.T(), uint64(2), d.GetWorkerCount())

	d.DecWorker()
	assert.Equal(s.T(), uint64(1), d.GetWorkerCount())
}

func (s *Suit) TestMaxWorkers() {
	d := dispatcher.NewDispatcher()
	assert.Equal(s.T(), uint64(0), d.GetMaxWorkers())

	d.UpdateMaxWorkers(100)
	assert.Equal(s.T(), uint64(100), d.GetMaxWorkers())

	d.UpdateMaxWorkers(1)
	assert.Equal(s.T(), uint64(1), d.GetMaxWorkers())
}

func (s *Suit) TestDispatch() {
	d := dispatcher.NewDispatcher(dispatcher.WithMaxWorkers(3))
	d.Dispatch()
	select {
	case ready, ok := <-d.WaitReady():
		assert.Equal(s.T(), true, ok)
		assert.Equal(s.T(), struct{}{}, ready)
	default:
	}

	d.IncWorker()
	d.IncWorker()
	d.IncWorker()
	d.Dispatch()
	select {
	case <-d.WaitReady():
	default:
		assert.EqualValues(s.T(), d.GetMaxWorkers(), d.GetWorkerCount())
	}

	err := d.OnShutdown()
	assert.NoError(s.T(), err)

	d.Dispatch()
	select {
	case ready, ok := <-d.WaitReady():
		assert.Equal(s.T(), false, ok)
		assert.Equal(s.T(), struct{}{}, ready)
	default:
	}
}
