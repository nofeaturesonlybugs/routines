package routines_test

import (
	"fmt"

	"github.com/nofeaturesonlybugs/routines"
)

// AlreadyStartedService shows how to use the Service and Routines interfaces.
type AlreadyStartedService struct {
	C <-chan int
	routines.Service
}

// NewAlreadyStartedService creates an instance of our service.
func NewAlreadyStartedService() *AlreadyStartedService {
	rv := &AlreadyStartedService{}
	rv.Service = routines.NewService(rv.start)
	return rv
}

// start starts the sample service.
func (me *AlreadyStartedService) start(routines routines.Routines) error {
	fmt.Println("AlreadyStartedService.start")
	defer fmt.Println("AlreadyStartedService.start returned")

	syncCh := make(chan struct{}, 1)
	routines.Go(func() {
		fmt.Println("goroutine start")
		defer fmt.Println("goroutine returned")

		c := make(chan int, 3)
		me.C = c
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
	<-syncCh
	return nil
}

func Example_alreadyStartedService() {
	fmt.Println("main start")
	defer fmt.Println("main returned")

	routines := routines.NewRoutines()
	defer fmt.Println("wait done")
	defer routines.Wait()
	defer fmt.Println("waiting")

	service := NewAlreadyStartedService()
	err := service.Start(routines)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer fmt.Println("AlreadyStartedService stopped")
	defer service.Stop()
	defer fmt.Println("AlreadyStartedService stopping")

	// Starting the service again results in an expected error!
	err = service.Start(routines)
	if err != nil {
		fmt.Println("Second start failed!")
	}

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
	// AlreadyStartedService.start
	// goroutine start
	// AlreadyStartedService.start returned
	// Second start failed!
	// 1
	// 2
	// 3
	// AlreadyStartedService stopping
	// goroutine returned
	// AlreadyStartedService stopped
	// waiting
	// wait done
	// main returned
}
