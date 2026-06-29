package commands

import (
	"testing"
)

func TestE2EConfigDoctor(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "config", "doctor")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !cr.hitPath("/server") {
		t.Errorf("doctor should probe server scope GET /server; paths = %v", cr.Paths)
	}
	if !cr.hitPath("/servers") {
		t.Errorf("doctor should probe account scope GET /servers; paths = %v", cr.Paths)
	}
}
