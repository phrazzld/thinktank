---
derived_from: simplicity
enforced_by: code review & race detector
id: concurrency-patterns
last_modified: '2025-05-14'
version: '0.1.0'
---
# Binding: Use Goroutines and Channels Judiciously with Explicit Coordination

Implement Go concurrency using clear patterns that prevent leaks, race conditions, and
deadlocks. Use goroutines only when they genuinely simplify design or improve
performance, pass context for propagating cancellation, use channels with clear
ownership semantics, and apply synchronization primitives correctly when sharing memory.

## Rationale

This binding directly implements our simplicity tenet by protecting against one of the
most challenging sources of complexity in software — uncontrolled concurrency. When
concurrent operations lack clear coordination patterns, they create an explosion of
potential execution orders that are impossible to reason about and test exhaustively.

Think of goroutine management like an organizational system for a team of workers.
Without clear communication protocols and coordination mechanisms, a team quickly
devolves into chaos—some workers might continue on obsolete tasks, others might work in
isolation forever, and still others might fight over shared resources. Similarly,
unmanaged goroutines can leak, waste resources on canceled tasks, or create subtle data
races. By establishing explicit coordination patterns through channels, context, and
synchronization primitives, you create a predictable, understandable concurrent system,
just as clear protocols create a well-functioning team.

The apparent simplicity of Go's concurrency mechanisms—the ease of spinning up
goroutines and creating channels—belies the potential complexity they introduce. Each
`go` statement creates a potential source of non-determinism in your program, an
invisible execution path that might interact with shared state in unexpected ways. This
hidden complexity compounds with each additional goroutine, potentially creating systems
where bugs are intermittent, difficult to reproduce, and frustratingly subtle.
Disciplined use of Go's concurrency features, guided by established patterns, prevents
this complexity explosion before it starts, keeping your systems simple even as they
leverage the power of concurrent execution.

## Rule Definition

This binding establishes clear requirements for implementing concurrency in Go:

- **Goroutine Management**:

  - Goroutines MUST NOT be started without a clear strategy for termination
  - Every goroutine MUST have at least one of:
    - A defined exit condition that is guaranteed to occur
    - A cancellation mechanism through context
    - A done/quit channel for explicit shutdown signaling
  - Goroutines that start other goroutines are responsible for their cleanup
  - Use `go` keyword only when concurrency is necessary, not just to run code
    asynchronously
  - Worker pools SHOULD be used for bounding concurrent operations rather than unbounded
    goroutine creation

- **Context Usage**:

  - Functions that perform I/O, long-running operations, or spawn goroutines MUST accept
    a `context.Context` parameter
  - Context MUST be the first parameter of a function that accepts it
  - Context cancellation MUST be respected by checking `ctx.Done()` in long-running
    operations
  - Context values should be used sparingly, primarily for request-scoped data like
    trace IDs
  - DO NOT store essential function parameters in context; use explicit parameters
    instead
  - Context MUST be propagated through call chains to lower-level functions

- **Channel Patterns**:

  - Channels MUST have clear ownership semantics:
    - Exactly one goroutine is responsible for closing each channel
    - Senders typically must not close channels they don't own
    - Use close for signaling completion, not special values
  - Channel direction (send/receive) SHOULD be specified in function parameters
  - Unbuffered channels SHOULD be preferred for synchronization guarantees
  - Buffered channels SHOULD be used with clear buffer size justification
  - The `select` statement MUST include a cancellation case for potentially blocking
    operations
  - Avoid complex channel-of-channels patterns unless unavoidable

- **Synchronization Primitives**:

  - Use `sync.Mutex` or `sync.RWMutex` when data must be shared between goroutines
  - Always release locks in a defer statement when possible to prevent leaks
  - Keep the critical section (locked code) as small as possible
  - Use `sync.WaitGroup` for waiting on groups of goroutines to complete
  - Use `sync.Once` for one-time initialization that must happen exactly once
  - Prefer atomic operations from `sync/atomic` for simple counters and flags
  - DO NOT copy mutex-containing structs; use pointers instead

- **Race Safety**:

  - All code MUST pass the race detector (`go test -race ./...`)
  - All shared data access across goroutines MUST be properly synchronized
  - Avoid subtle data races with maps, slices, and interface values
  - BE AWARE of Go's memory model which makes specific guarantees about when writes
    become visible to other goroutines
  - Tests for concurrent code SHOULD be run multiple times to increase chances of
    catching race conditions

## Practical Implementation

1. **Structured Goroutine Management**: Implement clear patterns for goroutine lifetime
   control:

   ```go
   // ✅ GOOD: Worker pool with controlled concurrency
   func ProcessItems(ctx context.Context, items []Item) error {
       // Create a bounded pool of workers
       const maxWorkers = 5
       sem := make(chan struct{}, maxWorkers)
       errCh := make(chan error, 1) // First error channel pattern
       done := make(chan struct{})

       var wg sync.WaitGroup

       // Start a goroutine to close done channel when all work completes
       go func() {
           wg.Wait()
           close(done)
       }()

       // Process items with bounded concurrency
       for i, item := range items {
           // Acquire semaphore slot or exit if context canceled
           select {
           case sem <- struct{}{}:
               // Continue with processing
           case <-ctx.Done():
               return ctx.Err()
           case err := <-errCh:
               // Return first error encountered
               return err
           }

           wg.Add(1)
           go func(i int, item Item) {
               defer wg.Done()
               defer func() { <-sem }() // Release semaphore in any case

               if err := processItem(ctx, item); err != nil {
                   // Try to send error, but don't block if one is already sent
                   select {
                   case errCh <- err:
                   default:
                   }
               }
           }(i, item)
       }

       // Wait for either completion or context cancellation
       select {
       case <-done:
           return nil
       case <-ctx.Done():
           return ctx.Err()
       case err := <-errCh:
           return err
       }
   }
   ```

1. **Effective Context Propagation**: Use context correctly for cancellation and
   timeout:

   ```go
   // ✅ GOOD: Proper context usage with timeout
   func FetchUserData(ctx context.Context, userID string) (*UserData, error) {
       // Create a timeout for this specific operation
       timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
       defer cancel() // Always cancel to release resources

       // Make HTTP request with context
       req, err := http.NewRequestWithContext(timeoutCtx, "GET",
           fmt.Sprintf("https://api.example.com/users/%s", userID), nil)
       if err != nil {
           return nil, fmt.Errorf("creating request: %w", err)
       }

       resp, err := http.DefaultClient.Do(req)
       if err != nil {
           // Check specifically for timeout/cancellation
           if timeoutCtx.Err() != nil {
               return nil, fmt.Errorf("request timed out: %w", timeoutCtx.Err())
           }
           return nil, fmt.Errorf("making request: %w", err)
       }
       defer resp.Body.Close()

       // Use a separate defer to avoid holding lock during I/O
       var data UserData
       if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
           return nil, fmt.Errorf("decoding response: %w", err)
       }

       return &data, nil
   }

   // ✅ GOOD: Check for cancellation in loops and long operations
   func ProcessLargeDataset(ctx context.Context, dataset []Data) error {
       for i, data := range dataset {
           // Check for cancellation regularly in loops
           if i%100 == 0 {
               select {
               case <-ctx.Done():
                   return ctx.Err()
               default:
                   // Continue processing
               }
           }

           if err := processData(ctx, data); err != nil {
               return err
           }
       }
       return nil
   }
   ```

1. **Channel Ownership and Direction**: Establish clear ownership semantics:

   ```go
   // ✅ GOOD: Clear sender/receiver responsibilities with direction

   // Generator owns and closes the returned channel
   func generateItems(ctx context.Context) <-chan Item {
       ch := make(chan Item)

       go func() {
           defer close(ch) // Generator owns the channel and is responsible for closing

           for i := 0; i < 10; i++ {
               select {
               case <-ctx.Done():
                   return
               case ch <- Item{ID: i}:
                   // Item sent successfully
               }
           }
       }()

       return ch // Return receive-only channel
   }

   // Processor accepts a receive-only channel, doesn't close it
   func processItems(ctx context.Context, items <-chan Item) <-chan Result {
       results := make(chan Result)

       go func() {
           defer close(results) // Processor owns results channel and closes it

           for item := range items {
               result := process(item)

               select {
               case <-ctx.Done():
                   return
               case results <- result:
                   // Result sent successfully
               }
           }
       }()

       return results
   }

   // Main orchestration
   func main() {
       ctx, cancel := context.WithCancel(context.Background())
       defer cancel()

       items := generateItems(ctx)
       results := processItems(ctx, items)

       // Consume results
       for result := range results {
           fmt.Println(result)
       }
   }
   ```

1. **Proper Synchronization Primitives**: Use the right tool for each synchronization
   need:

   ```go
   // ✅ GOOD: Appropriate mutex usage with defer
   type Cache struct {
       mu      sync.RWMutex
       entries map[string]Entry
   }

   func (c *Cache) Get(key string) (Entry, bool) {
       c.mu.RLock() // Read lock for concurrent reads
       defer c.mu.RUnlock()

       entry, found := c.entries[key]
       return entry, found
   }

   func (c *Cache) Set(key string, entry Entry) {
       c.mu.Lock() // Write lock for exclusive access
       defer c.mu.Unlock()

       if c.entries == nil {
           c.entries = make(map[string]Entry)
       }
       c.entries[key] = entry
   }

   // ✅ GOOD: WaitGroup for coordinating multiple goroutines
   func processInParallel(items []Item) []Result {
       results := make([]Result, len(items))
       var wg sync.WaitGroup

       for i, item := range items {
           wg.Add(1)
           go func(i int, item Item) {
               defer wg.Done()
               results[i] = processItem(item)
           }(i, item)
       }

       wg.Wait() // Wait for all goroutines to finish
       return results
   }

   // ✅ GOOD: Using Once for safe one-time initialization
   var (
       instance *Singleton
       once     sync.Once
   )

   func GetInstance() *Singleton {
       once.Do(func() {
           instance = &Singleton{}
           instance.Initialize()
       })
       return instance
   }
   ```

1. **Race Detection and Testing**: Implement thorough concurrent testing:

   ```go
   // ✅ GOOD: Testing concurrent access
   func TestCacheThreadSafety(t *testing.T) {
       cache := NewCache()
       const iterations = 1000

       var wg sync.WaitGroup
       wg.Add(2) // One writer, one reader

       // Writer goroutine
       go func() {
           defer wg.Done()
           for i := 0; i < iterations; i++ {
               key := fmt.Sprintf("key-%d", i%100) // Create key collisions
               cache.Set(key, fmt.Sprintf("value-%d", i))
           }
       }()

       // Reader goroutine
       go func() {
           defer wg.Done()
           for i := 0; i < iterations; i++ {
               key := fmt.Sprintf("key-%d", i%100)
               _, _ = cache.Get(key) // Just exercise the get path
           }
       }()

       wg.Wait()
   }

   // Run with:
   // go test -race -count=5 ./...
   ```

## Examples

```go
// ❌ BAD: Goroutine leak without cancellation
func StartWorker() {
    go func() {
        for {
            // Process work indefinitely with no way to stop
            time.Sleep(1 * time.Second)
            doWork()
        }
    }()
    // No way to stop the goroutine - it will run forever
}

// ✅ GOOD: Proper lifecycle management with cancellation
func StartWorker(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                // Clean shutdown when context is canceled
                return
            case <-ticker.C:
                doWork()
            }
        }
    }()
}

// Usage
ctx, cancel := context.WithCancel(context.Background())
StartWorker(ctx)
// Later, when the worker should stop:
cancel()
```

```go
// ❌ BAD: Channel with unclear ownership
func processData(data []int) {
    results := make(chan int)

    // Start workers to process data
    for _, d := range data {
        go func(d int) {
            results <- process(d)
        }(d)
    }

    // Read some results
    for i := 0; i < len(data)/2; i++ {
        fmt.Println(<-results)
    }

    // PROBLEM: Channel is never closed, remaining goroutines block forever
    // PROBLEM: Only reads half the results, causing goroutine leak
}

// ✅ GOOD: Clear channel ownership and complete consumption
func processData(data []int) []int {
    results := make(chan int, len(data)) // Buffered to avoid blocking
    var wg sync.WaitGroup

    // Start workers to process data
    for _, d := range data {
        wg.Add(1)
        go func(d int) {
            defer wg.Done()
            results <- process(d)
        }(d)
    }

    // Close channel when all workers are done
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect all results
    var collected []int
    for r := range results {
        collected = append(collected, r)
    }

    return collected
}
```

```go
// ❌ BAD: Race condition with shared data
type Counter struct {
    Value int
}

func (c *Counter) Increment() {
    c.Value++ // Unsynchronized access
}

func UseCounter() {
    counter := &Counter{}

    // Start 100 goroutines that all increment the counter
    for i := 0; i < 100; i++ {
        go func() {
            counter.Increment()
        }()
    }

    time.Sleep(time.Second)
    // Value will likely not be 100 due to race conditions
    fmt.Println(counter.Value)
}

// ✅ GOOD: Proper synchronization with mutex
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

func UseCounter() {
    counter := &Counter{}
    var wg sync.WaitGroup

    // Start 100 goroutines that all increment the counter
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }

    wg.Wait()
    // Value will be exactly 100
    fmt.Println(counter.Value())
}
```

```go
// ❌ BAD: Select without cancellation
func fetchData(urls []string) []Result {
    ch := make(chan Result)

    for _, url := range urls {
        go func(url string) {
            // This might block forever if network hangs
            data, err := http.Get(url)
            ch <- Result{Data: data, Err: err}
        }(url)
    }

    // Collect results, but no way to time out or cancel
    var results []Result
    for range urls {
        results = append(results, <-ch)
    }
    return results
}

// ✅ GOOD: Select with timeout and cancellation
func fetchData(ctx context.Context, urls []string) ([]Result, error) {
    ch := make(chan Result)
    var wg sync.WaitGroup

    for _, url := range urls {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()

            // Create request with context for cancellation
            req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
            if err != nil {
                select {
                case ch <- Result{Err: err}:
                case <-ctx.Done():
                }
                return
            }

            resp, err := http.DefaultClient.Do(req)
            select {
            case ch <- Result{Data: resp, Err: err}:
            case <-ctx.Done():
                if resp != nil {
                    resp.Body.Close() // Prevent resource leak
                }
            }
        }(url)
    }

    // Close channel when all fetches complete
    go func() {
        wg.Wait()
        close(ch)
    }()

    // Collect results with timeout
    var results []Result
    for {
        select {
        case result, ok := <-ch:
            if !ok {
                return results, nil // Channel closed, all done
            }
            results = append(results, result)
        case <-ctx.Done():
            return results, ctx.Err()
        }
    }
}
```

## Related Bindings

- [go-error-wrapping](../../docs/bindings/categories/go/error-wrapping.md): Proper error handling is even more
  important in concurrent code, where errors can occur in multiple goroutines
  simultaneously. This binding provides patterns for wrapping errors with context, which
  is essential for debugging issues in concurrent systems where the source of an error
  might be in a different goroutine than where it's observed.

- [go-interface-design](../../docs/bindings/categories/go/interface-design.md): Well-designed interfaces make
  concurrent code more testable and maintainable by providing clear abstraction
  boundaries. Abstractions like `io.Reader` and `io.Writer` enable concurrent operations
  by defining clear contracts for how components interact, which is essential for safe
  concurrency.

- [go-package-design](../../docs/bindings/categories/go/package-design.md): Good package organization helps control how
  concurrent code is structured and exposed. By defining clear package boundaries and
  responsibilities, you can ensure that concurrency primitives are encapsulated
  appropriately and not exposed unnecessarily.

- [pure-functions](../../docs/bindings/core/pure-functions.md): Pure functions without side effects are
  inherently thread-safe and can be executed concurrently without synchronization. By
  maximizing the use of pure functions in your code, you reduce the need for complex
  concurrency controls and minimize the risk of race conditions.

- [immutable-by-default](../../docs/bindings/core/immutable-by-default.md): Immutable data structures eliminate
  an entire class of concurrency bugs by preventing shared state from being modified.
  When data is immutable, multiple goroutines can safely access it without
  synchronization, simplifying concurrent code substantially.
