package util

import (
	"bufio"
	"os/exec"
)

// ExecAndLog 执行命令并实时输出日志
func ExecAndLog(name string, args []string) error {
	cmd := exec.Command(name, args...)
	Info("will execute: %s", cmd.String())

	outScanner, errScanner, err := createScanner(cmd)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		for errScanner.Scan() {
			Info(errScanner.Text() + "\n")
		}
	}()

	go func() {
		for outScanner.Scan() {
			Info(outScanner.Text() + "\n")
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
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
