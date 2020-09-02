package routines

import (
	"testing"
)

func TestService_start_nilReceiver(t *testing.T) {
	var svc Service
	var ptr *service
	svc = ptr

	err := svc.Start(NewRoutines())
	if err == nil {
		t.FailNow()
	}
}
