# go-queue

## Usage
### func WithContext
```go
func WithContext(ctx context.Context)
```

### func WithQueueSize
```go
func WithQueueSize(size uint64)
```

### func WithMaxWorkers
```go
func WithMaxWorkers(maxCount uint64)
```

## examples
```go
q := queue.NewQueue()

go func() {
    for err := range q.Errors() {
        log.Printf("error: %s\n", err.Error())
    }
}()

q.AddJob(func(ctx context.Context) error {
    log.Print("job 001")
    time.Sleep(3 * time.Second)
    log.Print("job 001 finish")
    return nil
})

q.AddJob(func(ctx context.Context) error {
    log.Print("job 002")
    time.Sleep(5 * time.Second)
    log.Print("job 002 finish")
    return nil
})

q.AddJob(func(ctx context.Context) error {
    log.Print("job 003 default retry")
    return errors.New("job 003 finish")
}, queue.MaxRetries(1), queue.WithBackoff(queue.BackoffLinear(0*time.Second)))

time.Sleep(1 * time.Second)

q.AddJob(func(ctx context.Context) error {
    log.Print("job 004")
    time.Sleep(1 * time.Second)
    log.Print("job 004 finish")
    return nil
})

q.AddJob(func(ctx context.Context) error {
    log.Print("job 005 custom retry")
    time.Sleep(1 * time.Second)
    return errors.New("job 005 finish")
}, queue.MaxRetries(3), queue.WithBackoff(queue.BackoffExponential(1*time.Second)))

<-q.Done()
```