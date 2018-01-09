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
	d := &dmon.Dmon{EventEmitter: eventEmitter, ProcessManager: processManager}

	if err := d.CheckFilesystemAvailability(logger, dirToCheck, time.Second*10); err != nil {
		os.Exit(1)
	}
}

func printUsageAndExit() {
	fmt.Println("usage: dmon --metron-endpoint <metron endpoint> <directory to check>")
	os.Exit(1)
}
