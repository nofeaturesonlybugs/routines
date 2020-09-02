package routines

import (
	"testing"
)

func TestRoutines_nil(t *testing.T) {
	// Tests some functions when using a nil Routines; part of achieving 100% test coverage.
	var r *routines

	if r.Done() != nil {
		t.FailNow()
	}

	if r.Child() != nil {
		t.FailNow()
	}
}

func TestRoutines_Stop_multipleTimes(t *testing.T) {
	// Tests that calling Stop() multiple times does not cause a panic.
	r := NewRoutines()
	r.Stop()
	r.Stop()
	r.Stop()
}

func TestRoutines_Stop_parentStopsChildren(t *testing.T) {
	p := NewRoutines()
	c := p.Child()

	defer c.Wait()
	p.Stop()
}
