package routines_test

import (
	"fmt"

	"github.com/nofeaturesonlybugs/routines"
)

// SampleService shows how to use the Service and Routines interfaces.
type SampleService struct {
	C <-chan int
	routines.Service
}

// NewSampleService creates an instance of our service.
func NewSampleService() *SampleService {
	rv := &SampleService{}
	rv.Service = routines.NewService(rv.start)
	return rv
}

// start starts the sample service.
func (me *SampleService) start(routines routines.Routines) error {
	fmt.Println("SampleService.start")
	defer fmt.Println("SampleService.start returned")

	// syncCh is used to prevent start() from returning until me.C has been set below; that
	// is the point at which this service is considered "ready."
	syncCh := make(chan struct{}, 1)
	routines.Go(func() {
		fmt.Println("goroutine start")
		defer fmt.Println("goroutine returned")

		c := make(chan int, 3)
		me.C = c
		// With me.C assigned we can safely return from start(); closing the channel will do so.
		close(syncCh)
		c <- 1
		c <- 2
		c <- 3
		for {
			select {
			case <-routines.Done():
				goto done
			}
		}
	done:
		me.C = nil
	})
	// Waiting on the channel to be closed before returning.
	<-syncCh
	return nil
}

func Example_sampleService() {
	fmt.Println("main start")
	defer fmt.Println("main returned")

	routines := routines.NewRoutines()
	defer fmt.Println("wait done")
	defer routines.Wait()
	defer fmt.Println("waiting")

	service := NewSampleService()
	err := service.Start(routines)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
	// SampleService.start
	// goroutine start
	// SampleService.start returned
	// 1
	// 2
	// 3
	// SampleService stopping
	// goroutine returned
	// SampleService stopped
	// waiting
	// wait done
	// main returned
}
