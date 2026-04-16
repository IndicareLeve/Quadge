package system

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func StreamLogs(name string) (io.ReadCloser, error) {
	cmd := exec.Command("journalctl", "--user", "-u", name+".service", "-f", "-n", "50")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting journalctl: %w", err)
	}

	return &logReader{
		stdout: stdout,
		cmd:    cmd,
	}, nil
}

type logReader struct {
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (lr *logReader) Read(p []byte) (n int, err error) {
	return lr.stdout.Read(p)
}

func (lr *logReader) Close() error {
	if lr.cmd.Process != nil {
		lr.cmd.Process.Kill()
	}
	return lr.stdout.Close()
}

func FormatLogLine(line string) string {
	return line
}

func NewLogScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	return scanner
}
