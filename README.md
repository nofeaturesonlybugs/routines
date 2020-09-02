[![Documentation](https://godoc.org/github.com/nofeaturesonlybugs/routines?status.svg)](http://godoc.org/github.com/nofeaturesonlybugs/routines)
[![Go Report Card](https://goreportcard.com/badge/github.com/nofeaturesonlybugs/routines)](https://goreportcard.com/report/github.com/nofeaturesonlybugs/routines)
[![Build Status](https://travis-ci.com/nofeaturesonlybugs/routines.svg?branch=master)](https://travis-ci.com/nofeaturesonlybugs/routines)
[![codecov](https://codecov.io/gh/nofeaturesonlybugs/routines/branch/master/graph/badge.svg)](https://codecov.io/gh/nofeaturesonlybugs/routines)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Routines
Golang package for enhanced goroutine synchronization and well defined idiomatic Golang services.

# Dependencies
* github.com/nofeaturesonlybugs/errors

# Write Idiomatic Client Code
`bus` and `listen` are considered mock packages for this example.
```go
func fatal(e err) {
    if e != nil {
        fmt.Println("Error", e)
        os.Exit(255)
    }
}

func main() {
    // Top-level Routines object.
    rtns := routines.NewRoutines()
    // WaitGroup-like behavior providing clean shutdown.
    defer rtns.Wait()

    bus := bus.NewBus()
    err := bus.Start(rtns)
    fatal(err)
    defer bus.Stop()

    listener, err := listen.New("localhost:8080")
    fatal(err)
    err = listener.Start(rtns)
    fatal(err)
    defer listener.Stop()

    // Program stalls waiting on sigc or other shutdown signal...
}
```
The preceding example demonstrates the following well-defined behaviors:
* Calls to Start() block until the service is fully started, preventing race conditions
    * Returns an error if the service can not start
* Calls to Stop() block until the service is fully stopped, preventing race conditions

# Write Lean Services
The `Service` interface provides the well-defined behavior of `Start` and `Stop` without polluting your service implementation.
```go
type MyService struct {
    Routines.Service
}

func NewMyService() *MyService {
    rv := &MyService{}
    rv.Service = routines.NewService(rv.start)
    return rv
}

func (me *MyService) start(rtns routines.Routines) error {
    listener, err := net.Listen("localhost:8080")
    if err != nil {
        return err
    }
    closed := make(chan struct{}, 1)

    // Create a lambda function that handls the connection; we'll pass the
    // returned function to rtns.Go() to ensure all handlers are finished
    // when the service stops.
    handler := func(c net.Conn) func() {
        return func() {
            io.Copy(c, c)
            c.Close()
        }
    }
    // Continuous loop that accepts and handles connections.
    accept := func() {
        for {
            if conn, err := listener.Accept(); err != nil {
                select {
                    case <-closed:
                        goto done
                    default:
                        // Handle error
                }
            } else {
                rtns.Go(handler(conn))
            }
        }
        done:
    }
    // When rtns.Done() signals we close the listener; this ends the accept
    // function.
    cleanup := func() {
        <- rtns.Done()
        close(closed)
        listener.Close()
    }
    rtns.Go(cleanup)
    rtns.Go(accept)
    return nil
}
```

The preceding example is not concerned with the concurrency primitives required to implement correct `Start` and `Stop` behavior; it is only concerned with correctly implementing the service behavior.

All goroutines created by the service are invoked with `rtns.Go()`.  This ensures clean shutdown when the client calls `Stop()` as `Stop()` will not return until all such goroutines have completed.

Also notice that the `struct` itself is not storing any variables or resources created as part of starting the service.  Such resources are instantiated in the call to `start()` and then remembered by the goroutines of the service.  When the service is stopped all of these goroutines will end; the handles to the resources will disappear and they will be garbage collected.  There is no need to remember to set such struct members to nil.