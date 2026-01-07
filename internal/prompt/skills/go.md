
# Go Idioms

Guidelines for writing idiomatic, maintainable Go code.

## When to Activate

- Writing or editing Go code (`.go` files)
- Creating new Go packages
- Reviewing Go code

## Error Handling

### Always handle errors explicitly

```go
// GOOD
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}

// BAD - silent error
result, _ := doSomething()
```

### Wrap errors with context

```go
// GOOD - adds context
if err := db.Query(sql); err != nil {
    return fmt.Errorf("querying users: %w", err)
}

// BAD - loses context
if err := db.Query(sql); err != nil {
    return err
}
```

### Use sentinel errors for expected conditions

```go
var ErrNotFound = errors.New("not found")

func FindUser(id string) (*User, error) {
    user := cache.Get(id)
    if user == nil {
        return nil, ErrNotFound
    }
    return user, nil
}

// Caller can check
if errors.Is(err, ErrNotFound) {
    // handle not found case
}
```

## Naming

### Use short, descriptive names

```go
// GOOD
func (u *User) Name() string
func (c *Client) Do(req *Request) (*Response, error)
var buf bytes.Buffer

// BAD - overly verbose
func (user *User) GetUserName() string
func (httpClient *HTTPClient) ExecuteRequest(request *HTTPRequest) (*HTTPResponse, error)
var buffer bytes.Buffer
```

### Acronyms should be all caps or all lower

```go
// GOOD
type HTTPClient struct{}
var userID string
func parseURL(s string)

// BAD - mixed case
type HttpClient struct{}
var UserId string
```

## Structs

### Use struct literals with field names

```go
// GOOD - explicit, maintainable
user := User{
    Name:  "Alice",
    Email: "alice@example.com",
    Admin: true,
}

// BAD - positional, fragile
user := User{"Alice", "alice@example.com", true}
```

### Prefer composition over embedding

```go
// GOOD - explicit composition
type Server struct {
    logger *Logger
    db     *Database
}

func (s *Server) Log(msg string) {
    s.logger.Info(msg)
}

// CAREFUL WITH - embedding can expose too much
type Server struct {
    *Logger  // All Logger methods now on Server
    *Database
}
```

## Interfaces

### Define interfaces where they're used

```go
// GOOD - consumer defines interface
package order

type PaymentProcessor interface {
    Charge(amount int) error
}

type Service struct {
    payments PaymentProcessor
}

// BAD - producer defines interface
package payment

type Processor interface {
    Charge(amount int) error
}

type StripeProcessor struct{}
```

### Keep interfaces small

```go
// GOOD - single method
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Compose when needed
type ReadWriter interface {
    Reader
    Writer
}

// BAD - too many methods
type DataStore interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte) error
    Delete(key string) error
    List(prefix string) ([]string, error)
    Watch(key string) <-chan Event
    // ... etc
}
```

## Concurrency

### Use channels for communication, mutexes for state

```go
// GOOD - channel for signaling
done := make(chan struct{})
go func() {
    // work
    close(done)
}()
<-done

// GOOD - mutex for shared state
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

### Use context for cancellation

```go
func FetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}
```

## Testing

### Use table-driven tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", 
                    tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

### Use testify for assertions when helpful

```go
import "github.com/stretchr/testify/assert"

func TestUser(t *testing.T) {
    user, err := CreateUser("alice@example.com")
    
    assert.NoError(t, err)
    assert.Equal(t, "alice@example.com", user.Email)
    assert.True(t, user.Active)
}
```

## Packages

### Keep package names simple

```go
// GOOD
package http
package user
package order

// BAD
package httputils
package userservice
package orderhandlers
```

### Don't stutter

```go
// GOOD
http.Client
user.Service
order.Create

// BAD
http.HTTPClient
user.UserService
order.OrderCreate
```
