package dmon

import (
	"fmt"

	"code.cloudfoundry.org/dropsonde"
	"code.cloudfoundry.org/dropsonde/envelopes"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/sonde-go/events"
)

type LoggingEventEmitter struct {
	Logger lager.Logger
}

func (e *LoggingEventEmitter) EmitEvent() error {
	e.Logger.Info("badness-occurred")
	return nil
}

type MetronEventEmitter struct {
	MetronEndpoint string
	DirToCheck     string
}

func (e *MetronEventEmitter) EmitEvent() error {
	source := "dmon"

	if err := dropsonde.Initialize(e.MetronEndpoint, source); err != nil {
		return err
	}

	message := fmt.Sprintf("Filesystem containing %s unavailable", e.DirToCheck)

	errorEvent := events.Error{
		Source:  &source,
		Message: &message,
	}

	envelope := events.Envelope{
		Origin:    &source,
		EventType: events.Envelope_Error.Enum(),
		Error:     &errorEvent,
	}

	return envelopes.SendEnvelope(&envelope)
}
