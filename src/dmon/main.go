package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"dmon/dmon"

	"code.cloudfoundry.org/lager"
)

func main() {
	logger := lager.NewLogger("dmon")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	metronEndpoint := flag.String("metron-endpoint", "", "metron endpoint")
	debugProgram := flag.String("debug-program", "", "external data gathering program")
	debugDataDir := flag.String("debug-data-dir", "", "directory in which to store collected data")
	flag.Parse()

	if *metronEndpoint == "" {
		printUsageAndExit()
	}

	dirToCheck := flag.Arg(0)
	if dirToCheck == "" {
		printUsageAndExit()
	}

	eventEmitter := &dmon.MetronEventEmitter{
		MetronEndpoint: *metronEndpoint,
		DirToCheck:     dirToCheck,
	}
	processManager := &dmon.LinuxProcessManager{}
	dataCollector := &dmon.SystemDataCollector{ProcessManager: processManager, DebugProgram: *debugProgram, DebugDataDir: *debugDataDir}

	d := &dmon.Dmon{EventEmitter: eventEmitter, ProcessManager: processManager, DataCollector: dataCollector}

	if err := d.CheckFilesystemAvailability(logger, dirToCheck, time.Second*10); err != nil {
		os.Exit(1)
	}
}

func printUsageAndExit() {
	fmt.Println("usage: dmon --metron-endpoint <metron endpoint> [--debug-program <external debug program> --debug-data-dir <debug data directory>] <directory to check>")
	os.Exit(1)
}
