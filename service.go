package routines

import (
	"github.com/nofeaturesonlybugs/errors"
	"sync"
)

// Service is the interface for a long-lived process that can start once and later be stopped.
type Service interface {
	// Start starts the service.
	Start(Routines) error
	// Stop stops the service.
	Stop()
}

// service is the hidden internal type that implements the service interface.
type service struct {
	start    func(Routines) error
	mut      sync.Mutex
	routines Routines
}

// NewService creates a new service that will launch the start method only once until it
// is stopped.  If start is nil then a stub function will be created that returns an error.
func NewService(start func(Routines) error) Service {
	if start == nil {
		start = func(rtns Routines) error {
			return errors.NilArgument("start").Type(start)
		}
	}
	return &service{start: start}
}

// Start starts the service.
func (me *service) Start(routines Routines) error {
	if me == nil {
		return errors.NilReceiver().Type(me)
	}
	me.mut.Lock()
	defer me.mut.Unlock()
	if me.routines != nil {
		return errors.AlreadyStarted().Type(me)
	}

	var err error
	child := routines.Child()
	defer func() {
		if err != nil {
			child.Stop()
			child.Wait()
		}
	}()

	err = me.start(child)
	if err == nil {
		me.routines = child
	}

	return err
}

// Stop stops the service.
func (me *service) Stop() {
	if me != nil {
		me.mut.Lock()
		defer me.mut.Unlock()
		if me.routines != nil {
			me.routines.Stop()
			me.routines.Wait()
			me.routines = nil
		}
	}
}
