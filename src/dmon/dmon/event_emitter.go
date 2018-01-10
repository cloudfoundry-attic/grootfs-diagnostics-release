package dmon

import (
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/envelopes"
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
	arbitrary := int32(1)

	errorEvent := events.Error{
		Source:  &source,
		Code:    &arbitrary,
		Message: &message,
	}

	envelope := events.Envelope{
		Origin:    &source,
		EventType: events.Envelope_Error.Enum(),
		Error:     &errorEvent,
	}

	return envelopes.SendEnvelope(&envelope)
}
