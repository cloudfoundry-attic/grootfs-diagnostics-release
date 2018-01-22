package dmon

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/pkg/errors"
)

type Dmon struct {
	EventEmitter   EventEmitter
	DataCollector  DataCollector
	ProcessManager ProcessManager
}

//go:generate counterfeiter . EventEmitter
type EventEmitter interface {
	EmitEvent() error
}

//go:generate counterfeiter . DataCollector
type DataCollector interface {
	CollectData(logger lager.Logger)
}

//go:generate counterfeiter . ProcessManager
type ProcessManager interface {
	SpawnProcess(cmd *exec.Cmd) (int, error)
	Wait(pid int) (int, error)
}

const (
	timedOut                   = "timed-out"
	runningDiskCheckingProcess = "running-disk-checking-process"
	processExitedNonZero       = "process-exited-non-zero"
	spawning                   = "spawning"
	waiting                    = "waiting"
)

func (d *Dmon) CheckFilesystemAvailability(logger lager.Logger, dirToCheck string, writeTimeout time.Duration) error {
	logger = logger.Session("checking-fs-availability", lager.Data{"dir_to_check": dirToCheck})
	logger.Info("starting")
	defer logger.Info("finished")

	errs := func(err error, msg string, data lager.Data) error {
		logger.Error(msg, err, data)
		if eventErr := d.EventEmitter.EmitEvent(); eventErr != nil {
			logger.Error("emitting-event", eventErr)
		}
		if msg == timedOut {
			d.DataCollector.CollectData(logger)
		}
		return errors.Wrap(err, msg)
	}

	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("echo test > %s", filepath.Join(dirToCheck, "fs_availability_check")))

	exitChan := make(chan exitStatus)
	go d.spawnAndWait(cmd, exitChan)
	select {
	case waitStatus := <-exitChan:
		if waitStatus.err != nil {
			return errs(waitStatus.err, runningDiskCheckingProcess, nil)
		}

		if waitStatus.exitStatus != 0 {
			return errs(
				fmt.Errorf("expected exit status 0, got %d", waitStatus.exitStatus),
				processExitedNonZero, lager.Data{"exit_status": waitStatus.exitStatus},
			)
		}

		return nil

	case <-time.After(writeTimeout):
		return errs(fmt.Errorf("timed out after %dms", writeTimeout/time.Millisecond), timedOut, nil)
	}
}

type exitStatus struct {
	exitStatus int
	err        error
}

func (d *Dmon) spawnAndWait(cmd *exec.Cmd, exitChan chan<- exitStatus) {
	errs := func(err error, msg string) {
		exitChan <- exitStatus{err: errors.Wrap(err, msg)}
	}

	pid, err := d.ProcessManager.SpawnProcess(cmd)
	if err != nil {
		errs(err, spawning)
		return
	}
	exitCode, err := d.ProcessManager.Wait(pid)
	if err != nil {
		errs(err, waiting)
		return
	}

	exitChan <- exitStatus{exitStatus: exitCode}
}
