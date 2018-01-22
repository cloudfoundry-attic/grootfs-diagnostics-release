package dmon_test

import (
	"errors"
	"time"

	"dmon/dmon"
	"dmon/dmon/dmonfakes"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dmon", func() {
	var (
		eventEmitter   *dmonfakes.FakeEventEmitter
		dataCollector  *dmonfakes.FakeDataCollector
		processManager *dmonfakes.FakeProcessManager
		dirToCheck     = "/some/dir"
		logger         lager.Logger
		writeTimeout   time.Duration

		d *dmon.Dmon
	)

	BeforeEach(func() {
		eventEmitter = new(dmonfakes.FakeEventEmitter)
		dataCollector = new(dmonfakes.FakeDataCollector)

		processManager = new(dmonfakes.FakeProcessManager)
		processManager.SpawnProcessReturns(4, nil)

		logger = lagertest.NewTestLogger("dmon-tests")
		writeTimeout = time.Second

		d = &dmon.Dmon{
			EventEmitter:   eventEmitter,
			DataCollector:  dataCollector,
			ProcessManager: processManager,
		}
	})

	Describe("success", func() {
		JustBeforeEach(func() {
			Expect(d.CheckFilesystemAvailability(logger, dirToCheck, writeTimeout)).To(Succeed())
		})

		It("spawns a process that writes a file in the dir to check", func() {
			Expect(processManager.SpawnProcessCallCount()).To(Equal(1))
			cmd := processManager.SpawnProcessArgsForCall(0)
			Expect(cmd.Path).To(HaveSuffix("/bash"))
			Expect(cmd.Args).To(Equal([]string{"bash", "-c", "echo test > /some/dir/fs_availability_check"}))
		})

		It("waits on the returned pid", func() {
			Expect(processManager.WaitCallCount()).To(Equal(1))
			Expect(processManager.WaitArgsForCall(0)).To(Equal(4))
		})

		It("does not emit an event", func() {
			Expect(eventEmitter.EmitEventCallCount()).To(Equal(0))
		})

		It("does not collect any data", func() {
			Expect(dataCollector.CollectDataCallCount()).To(Equal(0))
		})
	})

	Describe("failures", func() {
		var checkErr error

		JustBeforeEach(func() {
			checkErr = d.CheckFilesystemAvailability(logger, dirToCheck, writeTimeout)
		})

		itEmitsEventAndReturnsError := func(errString string) {
			It("returns a wrapped error", func(done Done) {
				Expect(checkErr).To(MatchError(ContainSubstring(errString)))
				close(done)
			}, 1)

			It("emits an event", func(done Done) {
				Expect(eventEmitter.EmitEventCallCount()).To(Equal(1))
				close(done)
			}, 1)
		}

		Context("when the process spawning returns an error", func() {
			BeforeEach(func() {
				processManager.SpawnProcessReturns(0, errors.New("spawn-err"))
			})

			itEmitsEventAndReturnsError("spawn-err")

			It("does not collect data", func() {
				Expect(dataCollector.CollectDataCallCount()).To(Equal(0))
			})
		})

		Context("when the process exits with a non-zero exit code", func() {
			BeforeEach(func() {
				processManager.WaitReturns(1, nil)
			})

			itEmitsEventAndReturnsError("process-exited-non-zero")

			It("does not collect data", func() {
				Expect(dataCollector.CollectDataCallCount()).To(Equal(0))
			})
		})

		Context("when the process cannot be waited for", func() {
			BeforeEach(func() {
				processManager.WaitReturns(0, errors.New("wait-err"))
			})

			itEmitsEventAndReturnsError("wait-err")

			It("does not collect data", func() {
				Expect(dataCollector.CollectDataCallCount()).To(Equal(0))
			})
		})

		Context("when the process has not exited when the timeout is reached", func() {
			BeforeEach(func() {
				writeTimeout = time.Millisecond * 100
				processManager.WaitStub = func(_ int) (int, error) {
					time.Sleep(time.Second * 2)
					return 0, nil
				}
			})

			itEmitsEventAndReturnsError("timed out")

			It("collects data", func() {
				Expect(dataCollector.CollectDataCallCount()).To(Equal(1))
			})
		})
	})
})
