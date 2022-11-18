# go-queue

## Installation
```shell
go get github.com/fantasy9830/go-queue
```

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

### func WithBlocking
```go
// Blocking AddJob or Non-blocking AddJob
func WithBlocking(blocking bool)
```

## examples
```go
q := queue.NewQueue()

go func() {
    for err := range q.Errors() {
        log.Printf("error: %s\n", err.Error())
    }
}()

err := q.AddJob(func(ctx context.Context) error {
    log.Print("job 001")
    time.Sleep(3 * time.Second)
    log.Print("job 001 finish")
    return nil
})
if err != nil {
    log.Printf("add job 001 error: %s\n", err.Error())
}

err = q.AddJob(func(ctx context.Context) error {
    log.Print("job 002")
    time.Sleep(5 * time.Second)
    log.Print("job 002 finish")
    return nil
})
if err != nil {
    log.Printf("add job 002 error: %s\n", err.Error())
}

err = q.AddJob(func(ctx context.Context) error {
    log.Print("job 003 default retry")
    return errors.New("job 003 finish")
}, queue.MaxRetries(1), queue.WithBackoff(queue.BackoffLinear(0*time.Second)))
if err != nil {
    log.Printf("add job 003 error: %s\n", err.Error())
}

time.Sleep(1 * time.Second)

err = q.AddJob(func(ctx context.Context) error {
    log.Print("job 004")
    time.Sleep(1 * time.Second)
    log.Print("job 004 finish")
    return nil
})
if err != nil {
    log.Printf("add job 004 error: %s\n", err.Error())
}

err = q.AddJob(func(ctx context.Context) error {
    log.Print("job 005 custom retry")
    time.Sleep(1 * time.Second)
    return errors.New("job 005 finish")
}, queue.MaxRetries(3), queue.WithBackoff(queue.BackoffExponential(1*time.Second)))
if err != nil {
    log.Printf("add job 005 error: %s\n", err.Error())
}

<-q.Done()
```