package dmon

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/lager"
)

type SystemDataCollector struct {
	ProcessManager ProcessManager
	DebugProgram   string
	DebugDataDir   string
}

func (dc *SystemDataCollector) CollectData(logger lager.Logger) {
	logger.Session("collect-data")
	logger.Info("starting", lager.Data{"debug_program": dc.DebugProgram, "debug_data_dir": dc.DebugDataDir})
	defer logger.Info("ending")

	if dc.DebugProgram == "" || dc.DebugDataDir == "" {
		logger.Info("data collection not required")
		return
	}

	dc.executeExternalProgram(logger)
}

func (dc *SystemDataCollector) executeExternalProgram(logger lager.Logger) {
	dataTarball := filepath.Join(dc.DebugDataDir, fmt.Sprintf("dmon-debug-data-%d.tgz", time.Now().Unix()))
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s %s", dc.DebugProgram, dataTarball))

	pid, err := dc.ProcessManager.SpawnProcess(cmd)
	if err != nil {
		logger.Info(err.Error())
		return
	}
	_, err = dc.ProcessManager.Wait(pid)
	if err != nil {
		logger.Info(err.Error())
		return
	}
}
