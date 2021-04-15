package proxy

import (
	"go.temporal.io/sdk/workflow"
)

const (
	ResponseSignal = "completed"
)

type PayloadData struct {
	ByteData []byte
}

// Payload is the input for a signal execution, which will be transformed and passed down to the target as ProxySignalData
type Payload struct {
	PayloadData
	TargetId   string
	SignalName string
}

// SignalData is the input data for the signal. Any signal handler using this proxied method has to unmarshal to this received type
type SignalData struct {
	PayloadData
	CompletionTargetID string
}

// Result is the output of a signal execution, which will be returned on workflow completion
type Result struct {
	PayloadData
	Success bool
	Error   string
}

// RequestResponse is a workflow to coordinate a signal delivery and output check in a Request-Response fashion
func RequestResponse(ctx workflow.Context, payload *Payload) (*Result, error) {
	logger := workflow.GetLogger(ctx)
	completeChannel := workflow.GetSignalChannel(ctx, ResponseSignal)
	logger.Debug("Starting to proxy signal data", "targetId", payload.TargetId, "signalName", payload.SignalName)
	// wrap the target data
	wi := workflow.GetInfo(ctx)
	pd := SignalData{
		CompletionTargetID: wi.WorkflowExecution.ID,
		PayloadData: PayloadData{
			ByteData: payload.ByteData,
		},
	}
	// target the latest runID
	err := workflow.SignalExternalWorkflow(ctx, payload.TargetId, "", payload.SignalName, pd).Get(ctx, nil)
	if err != nil {
		logger.Error("Proxy signal failed", "error", err)
		return nil, err
	}
	// wait for completion
	var result *Result
	completeChannel.Receive(ctx, &result)
	logger.Debug("Finished proxying signal data", "targetId", payload.TargetId, "signalName", payload.SignalName)
	return result, nil
}
