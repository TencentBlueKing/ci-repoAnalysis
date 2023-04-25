package util

import (
	"os"
	"testing"
)

func TestExecAndLog(t *testing.T) {
	dir, _ := os.Getwd()
	cmd := "go"
	args := []string{"version"}
	err := ExecAndLog(cmd, args, dir)
	if err != nil {
		t.Fatalf("exec cmd %s failed: %s", cmd, err.Error())
	}
}
