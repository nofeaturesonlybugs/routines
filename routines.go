package routines

import "sync"

// Routines facilitates concurrency management between a program and its internal
// long-lived services.
type Routines interface {
	// Done returns a channel that will be closed when Stop() is called.
	Done() <-chan struct{}
	// Child creates a new Routines type as a child of the parent; calling Stop()
	// on the child does not affect the parent's routines but calling Stop()
	// on the parent will also stop all of the children.
	Child() Routines
	// Go launches the function as a go routine.
	Go(func())
	// Stop sends a stop signal to all routines started with Go().
	Stop()
	// Wait waits for all routines started with Go() to complete before returning.
	Wait()
}

// NewRoutines creates a routines type.
func NewRoutines() Routines {
	rv := &routines{
		doneCh: make(chan struct{})}
	return rv
}

// routines is the internal type that implement Routines interface.
type routines struct {
	doneCh chan struct{}
	//
	waitgroup sync.WaitGroup
	//
	parent *routines
	//
	children sync.WaitGroup
}

// childrenUp increments the childrenGroup WaitGroup by 1, is go routine safe, and nil pointer safe.
func (me *routines) childrenUp() {
	if me != nil {
		me.children.Add(1)
	}
}

// childrenDown decrements the childrenGroup WaitGroup, is go routine safe, and nil pointer safe.
func (me *routines) childrenDown() {
	if me != nil {
		me.children.Done()
	}
}

// Done returns a channel that will be closed when Stop() is called.  All functions
// started by Go() should end execution when this channel closes.
func (me *routines) Done() <-chan struct{} {
	if me == nil {
		return nil
	}
	return me.doneCh
}

// Child creates a new Routines type as a child of the parent; calling Stop()
// on the child does not affect the parent but calling Stop() on the parent will
// also stop all of the children.
func (me *routines) Child() Routines {
	if me == nil {
		return nil
	}
	rv := NewRoutines()
	rv.(*routines).parent = me
	//
	// Ensure the child is properly closed when the parent is stopped but also that the
	// child can stop early.
	fn := func() {
		parent, child := me, rv
		select {
		case <-parent.Done():
			// This signals parent.Stop() has been called; propagate the Stop() call to the child.
			child.Stop()
		case <-child.Done():
			// Stop() has been called on the child so no need to propagate the call any longer.
			goto done
		}
	done:
	}
	rv.Go(fn)
	//
	return rv
}

// Go launches the function as a go routine.  At a minimum `fn` must end execution
// when the channel returned by Done() is closed.
func (me *routines) Go(fn func()) {
	if me != nil {
		// Increment our own WaitGroup.
		me.waitgroup.Add(1)
		// Our parent contains a WaitGroup for all go routines launched by its children; increment that WaitGroup also.
		me.parent.childrenUp()
		go func() {
			defer me.waitgroup.Done()
			defer me.parent.childrenDown()
			fn()
		}()
	}
}

// Stop closes the channel returned by Done() and propagates the call to all child Routines
// created by Child().
func (me *routines) Stop() {
	if me != nil {
		// Closing a channel multiple times can panic.
		defer func() { recover() }()
		close(me.doneCh)
	}
}

// Wait waits for all routines started with Go() to complete before returning.
func (me *routines) Wait() {
	if me != nil {
		// First we wait on all of our children.
		me.children.Wait()
		// Then we wait on ourselfs.
		me.waitgroup.Wait()
	}
}
