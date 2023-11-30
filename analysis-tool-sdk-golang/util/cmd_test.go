package util

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestExecAndLog(t *testing.T) {
	dir, _ := os.Getwd()
	cmd := "go"
	args := []string{"version"}
	err := ExecAndLog(context.Background(), cmd, args, dir)
	if err != nil {
		t.Fatalf("exec cmd %s failed: %s", cmd, err.Error())
	}

	cmd = "sleep"
	args = []string{"5"}
	// test timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	err = ExecAndLog(ctx, cmd, args, dir)

	if err != nil {
		Info("exec cmd %s failed: %s", cmd, err.Error())
	}
	cancel()
}
