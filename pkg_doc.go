// Package routines provides enhanced goroutine synchronization.
//
// Dependencies
//
// The following packages act as dependencies:
//	github.com/nofeaturesonlybugs/errors
//
// Synopsis
//
// I wrote this package to facilitate a sane interface for types that act as long-lived
// services, consumers, or providers - henceforth referred to as services - within my Go programs.
//
// I consider the following a sane interface:
// 	type SomeService interface {
//		Start( Routines ) error
//		Stop()
//	}
//
// Furthermore a sane service is goroutine safe; Start() and Stop() can be called any number of times
// from any number of goroutines.  As noted below consecuitive calls to Start() should return an error.
//
// Start
//
// Start should have the following properties:
//	1. 	Blocks until the service is either started completely or can not start due to an error.
//	2. 	Starts and manages all of its own goroutines; goroutines should be considered operational
//		by the time Start() returns.
//	3.	Consecutive calls to Start() should return an error indicating the service is already started.
//	4.	Start requires a Routines type to perform synchronization with the rest of the program; many
//		go libraries use a context.Context here but it does not provide ordered shutdown.
//
// Stop
//
// Stop should have the following properties:
//	1.	Blocks until the service is stopped completely.
//
// Idiomatic Results
//
// Due to the properties of Start() and Stop() a service's invocation, error handling, and clean up will be:
//	routines := NewRoutines()
//	defer routines.Wait()
//
//	instance := NewService()
//	err := instance.Start( routines )
//	if err != nil {
//		// TODO
//	}
//	defer instance.Stop()
//
// The client code invoking the service is incredibly easy to read, follow, maintain, and update.
//
// The deferred Stop() will not return until the service is entirely done, which ensures a clean
// shutdown of the program.
//
// Any code following the above snippet can be assured the service is running and operational, which
// eliminates race conditions during program start up.
//
// Furthermore the client code is void of complexity; it is not directly handling the multiple
// primitives involved for ensuring the defined behavior.
//
// Concurrency Primitives
//
// To achieve the above requirements with the built-in concurrency primitives every service needs
// to manage a wait group, one or more synchronization channels, and possibly a context - or some
// combination of the three.
//
// The code implementing a service can quickly become cluttered with the code that manages its
// concurrency patterns, thus reducing readability and increasing complexity.  Reducing readability
// and increasing complexity have the combined effect of making the service difficult to maintain
// and difficult to enhance.
//
// The Routines type exported by this package separates the implementation of a service from the
// management of its concurrency - at least in regards to implementing correct Start() and Stop()
// methods as described above.
//
// A Sane Service
//
// The following is a faily bare implementation of a sane service using only the Routines interface;
// the next section further simplifies this implementation with the Service interface.
//
// The bare minimum struct definition:
//	type Sane struct {
//		// implementation details
//		...
//		// routines management
//		mut sync.Mutex
//		routines routines.Routines
//	}
//
// Only two properties are required as part of the structure definition.
//
// A sample Start() implementation:
//	func (me *Sane)Start( routines routines.Routines ) error {
//		if me == nil {
//			return fmt.Errorf("nil receiver")
//		}
//		me.mut.Lock()
//		defer me.mut.Unlock()
//		if me.routines != nil {
//			return fmt.Errorf("already started")
//		}
//
//		var err error
//
//		// Create a child but do not store it internally yet in case we exit with error.
//		child := routines.Child()
//		defer func() {
//			// If we exit with an error call child.Stop() to clean up the object.
//			if err != nil {
//				child.Stop()
//			}
//		}
//
//		err = // Some sort of initialization steps...
//		if err != nil {
//			return err
//		}
//
//		someResource, err := // More initialization...
//		if err != nil {
//			return err
//		}
//
//		// Start internal goroutine that uses someResource
//		child.Go( func(){
//			// child.Go manages the WaitGroup
//			for {
//				select {
//					case <-child.Done():
//						goto done
//
//					// either other cases or a default
//				}
//			}
//			done:
//		}())
//
//		// None of the above returned early with error; service is started; remember the child!
//		me.routines = child
//		return nil
//	}
//
// A sample Stop() implementation:
//	func (me *Sane)Stop() {
//		if me == nil {
//			return
//		}
//		me.mut.Lock()
//		defer me.mut.Unlock()
//		if me.routines != nil {
//			me.routines.Stop()
//			me.routines.Wait()
//			me.routines = nil
//		}
//	}
//
// Increasing Simplicity
//
// Both Start() and Stop() contain boiler plate and can be simplified further by introducing a
// new interface:
//	type Service interface {
//		Start( routines.Routines ) error
//		Stop()
//	}
//
// First let's make our service's Start() method private and remove some of the boiler plate:
//	func (me *Sane)start( routines routines.Routines ) error {
//		// Our new code structure will make it impossible for a nil receiver so technically we
//		// can remove this check.
//		// if me == nil {
//		//		return fmt.Errorf("nil receiver")
//		// }
//
//		// Our new interface will handle the mutex and the "already started" check.
//		// me.mut.Lock()
//		// defer me.mut.Unlock()
//		// if me.routines != nil {
//		// 	return fmt.Errorf("already started")
//		// }
//
//		var err error
//
//		// Our new interface will simplify this code block as well.
//		// Create a child but do not store it internally yet in case we exit with error.
//		// child := routines.Child()
//		// defer func() {
//		// 	// If we exit with an error call child.Stop() to clean up the object.
//		// 	if err != nil {
//		// 		child.Stop()
//		// 	}
//		// }
//
//		err = // Some sort of initialization steps...
//		if err != nil {
//			return err
//		}
//
//		someResource, err := // More initialization...
//		if err != nil {
//			return err
//		}
//
//		// Start internal goroutine that uses someResource
//		routines.Go( func(){
//			// routines.Go manages the WaitGroup
//			for {
//				select {
//					case <-routines.Done():
//						goto done
//
//					// either other cases or a default
//				}
//			}
//			done:
//		}())
//
//		// Our new interface will handle remembering of the child.
//		// None of the above returned early with error; service is started; remember the child!
//		// me.routines = child
//
//		return nil
//	}
//
// Here is the same method with the boilerplate removed; it is no longer concerned with its
// concurrency management and only concerned with implementing the service:
//	func (me *Sane)start( routines routines.Routines ) error {
//		err := // Some sort of initialization steps...
//		if err != nil {
//			return err
//		}
//
//		someResource, err := // More initialization...
//		if err != nil {
//			return err
//		}
//
//		// Start internal goroutine that uses someResource
//		routines.Go( func(){
//			// routines.Go manages the WaitGroup
//			for {
//				select {
//					case <-routines.Done():
//						goto done
//
//					// either other cases or a default
//				}
//			}
//			done:
//		}())
//		return nil
//	}
//
// The above start method is increadibly lean and deals only with the implementation details of the
// service - there is no more concurrency management for Start() and Stop().
//
// Next we will eliminate the Stop() method entirely; it's only purpose is to call Stop(), Wait(),
// and set the child routines object to nil.  This will all be handled by our Service interface.
//
// If you need to clean up any resources as part of stopping the service then the cleanest method is
// to do that after the `done` label in one of the service's goroutine methods created in start().
//
// Embed our Service interface into our service type:
//	type Sane struct {
//		// implementation details
//		...
//		Service
//	}
//
// Finally create a constructor function that sets the embedded Service correctly:
//	func NewSane() *Sane {
//		rv := &Sane{}
//		rv.Service = NewService( rv.start )
//		return rv
//	}
//
// The syntax required to set the embedded property is a little screwy; however it's a tradeoff worth
// making for consistently designed and implemented consumers, providers, or services within a larger
// program.
package routines
