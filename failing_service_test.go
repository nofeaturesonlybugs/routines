package routines_test

import (
	"fmt"

	"github.com/nofeaturesonlybugs/routines"
)

// FailingService shows a service failing to start.
type FailingService struct {
	C <-chan int
	routines.Service
}

// NewFailingService creates an instance of our service.
func NewFailingService() *FailingService {
	rv := &FailingService{}
	// Passing nil to NewService() will cause the call to Start() to return an error.
	rv.Service = routines.NewService(nil)
	return rv
}

func Example_failingService() {
	fmt.Println("main start")
	defer fmt.Println("main returned")

	routines := routines.NewRoutines()
	defer fmt.Println("wait done")
	defer routines.Wait()
	defer fmt.Println("waiting")

	service := NewFailingService()
	err := service.Start(routines)
	if err != nil {
		fmt.Println("Error starting service")
		return
	}
	defer fmt.Println("SampleService stopped")
	defer service.Stop()
	defer fmt.Println("SampleService stopping")

	count := 0
	for {
		select {
		case v := <-service.C:
			fmt.Println(v)
			// After 3 ints we set ch to nil so we'll print no more ints; we also call routines.Stop()
			// to shut down this function and the service.
			count++
			if count == 3 {
				goto done
			}
		}
	}
done:

	// Output: main start
	// Error starting service
	// waiting
	// wait done
	// main returned
}
