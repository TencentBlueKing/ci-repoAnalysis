package util

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ExecAndLog 执行命令并实时输出日志
func ExecAndLog(ctx context.Context, name string, args []string, workDir string) error {
	cmd := exec.CommandContext(ctx, name, args...)

	if len(workDir) > 0 {
		cmd.Dir = workDir
		Info("work directory: %s", workDir)
	}

	Info("will execute: %s", cmd.String())

	outScanner, errScanner, err := createScanner(cmd)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// 用于出错时上报最后[keepLine]行日志
	const keepLine = 10
	logs := [keepLine]string{}
	go func() {
		i := 0
		var line string
		for errScanner.Scan() {
			line = errScanner.Text()
			Info(line + "\n")
			logs[i%keepLine] = fmt.Sprintf("%d    : %s", i, line)
			i++
		}
	}()

	go func() {
		for outScanner.Scan() {
			Info(outScanner.Text() + "\n")
		}
	}()

	if err := cmd.Wait(); err != nil {
		errMsg := make([]string, 0, keepLine)
		for i := range logs {
			l := logs[i]
			if len(l) > 0 {
				errMsg = append(errMsg, l)
			}
		}
		errMsg = append(errMsg, err.Error())
		return errors.New(strings.Join(errMsg, "\n"))
	}
	return nil
}

func createScanner(cmd *exec.Cmd) (*bufio.Scanner, *bufio.Scanner, error) {
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	outScanner := bufio.NewScanner(outPipe)
	outScanner.Split(bufio.ScanLines)
	errScanner := bufio.NewScanner(errPipe)
	errScanner.Split(bufio.ScanLines)

	return outScanner, errScanner, nil
}
