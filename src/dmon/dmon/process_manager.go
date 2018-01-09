package dmon

import (
	"os"
	"os/exec"
	"syscall"
)

type LinuxProcessManager struct{}

func (p *LinuxProcessManager) SpawnProcess(cmd *exec.Cmd) (int, error) {
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	return cmd.Process.Pid, nil
}

func (p *LinuxProcessManager) Wait(pid int) (int, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, err
	}

	processState, err := proc.Wait()
	if err != nil {
		return 0, err
	}

	return processState.Sys().(syscall.WaitStatus).ExitStatus(), nil
}
