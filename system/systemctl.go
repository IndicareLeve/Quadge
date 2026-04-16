package system

import (
	"fmt"
	"os/exec"
	"strings"
)

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusFailed  ServiceStatus = "failed"
	StatusUnknown ServiceStatus = "unknown"
)

func runSystemctl(args ...string) (string, error) {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func GetServiceStatus(name string) (ServiceStatus, error) {
	output, err := runSystemctl("is-active", name+".service")
	output = strings.TrimSpace(output)

	if err == nil {
		switch output {
		case "active":
			return StatusRunning, nil
		case "inactive":
			return StatusStopped, nil
		case "failed":
			return StatusFailed, nil
		}
	}

	if strings.Contains(output, "inactive") || strings.Contains(output, "could not be found") {
		return StatusStopped, nil
	}

	return StatusUnknown, fmt.Errorf("checking status: %s", output)
}

func DaemonReload() error {
	_, err := runSystemctl("daemon-reload")
	if err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}
	return nil
}

func StartService(name string) error {
	_, err := runSystemctl("start", name+".service")
	if err != nil {
		return fmt.Errorf("starting service: %w", err)
	}
	return nil
}

func StopService(name string) error {
	_, err := runSystemctl("stop", name+".service")
	if err != nil {
		return fmt.Errorf("stopping service: %w", err)
	}
	return nil
}

func RestartService(name string) error {
	_, err := runSystemctl("restart", name+".service")
	if err != nil {
		return fmt.Errorf("restarting service: %w", err)
	}
	return nil
}
