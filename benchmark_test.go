package queue_test

import (
	"context"
	"testing"

	"github.com/fantasy9830/go-queue"
)

func BenchmarkQueueJob(b *testing.B) {
	b.ReportAllocs()
	q := queue.NewQueue()
	for n := 0; n < b.N; n++ {
		_ = q.AddJob(func(ctx context.Context) error {
			return nil
		})
	}
}
