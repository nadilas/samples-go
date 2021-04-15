package interactive_ui_signal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/temporalio/samples-go/interactive-ui-signal/proxy"
	"go.temporal.io/sdk/workflow"
)

func errorToProxy(ctx workflow.Context, ID string, err error) {
	workflow.SignalExternalWorkflow(ctx, ID, "", proxy.ResponseSignal, &proxy.Result{
		Error: err.Error(),
	})
}

func AccountWorkflow(ctx workflow.Context, account *Account) (*Account, error) {
	cleanHistory := false
	if account.Created == nil {
		ts := workflow.Now(ctx)
		account.Created = &ts
	}
	if account.Plan == "" {
		account.Plan = "trial"
	}

	// setup signals
	sel := workflow.NewSelector(ctx)
	logger := workflow.GetLogger(ctx)

	// upgrade request handler
	sel.AddReceive(workflow.GetSignalChannel(ctx, UpgradeAccountSignal), func(c workflow.ReceiveChannel, more bool) {
		var payload proxy.SignalData
		c.Receive(ctx, &payload)
		if payload.CompletionTargetID == "" {
			logger.Warn("Ignoring signal without completionID", "signal", UpgradeAccountSignal)
		}

		// either use raw ByteData to parse request
		var r *UpgradeRequest
		err := json.Unmarshal(payload.ByteData, &r)
		if err != nil {
			errorToProxy(ctx, payload.CompletionTargetID, err)
			return
		}

		// validate request
		if r.Actor != account.Name {
			errorToProxy(ctx, payload.CompletionTargetID, fmt.Errorf("access denied to account from %s", r.Actor))
			return
		}

		// execute update
		validFrom := workflow.Now(ctx)
		account.Plan = r.To
		account.PlanValidFrom = validFrom

		// send response using raw bytes
		resp := &UpgradeResponse{
			ValidFrom: validFrom,
		}
		bytes, err := json.Marshal(resp)
		if err != nil {
			errorToProxy(ctx, payload.CompletionTargetID, fmt.Errorf("failed to marshal response: %v", err))
			return
		}
		// send response
		workflow.SignalExternalWorkflow(
			ctx,
			payload.CompletionTargetID,
			"",
			proxy.ResponseSignal,
			&proxy.Result{
				Success: true,
				PayloadData: proxy.PayloadData{
					ByteData: bytes,
				},
			},
		)
	})

	// delete request handler
	sel.AddReceive(workflow.GetSignalChannel(ctx, DeleteAccountSignal), func(c workflow.ReceiveChannel, more bool) {
		var payload proxy.SignalData
		c.Receive(ctx, &payload)
		if payload.CompletionTargetID == "" {
			logger.Warn("Ignoring signal without completionID", "signal", UpgradeAccountSignal)
		}

		var r *DeleteAccountRequest
		err := json.Unmarshal(payload.ByteData, &r)
		if err != nil {
			errorToProxy(ctx, payload.CompletionTargetID, err)
			return
		}
		// validate request
		if r.Actor != account.Name {
			errorToProxy(ctx, payload.CompletionTargetID, fmt.Errorf("access denied to account from %s", r.Actor))
			return
		}

		// send response before workflow gets terminated
		_ = workflow.SignalExternalWorkflow(
			ctx,
			payload.CompletionTargetID,
			"",
			proxy.ResponseSignal,
			&proxy.Result{
				Success: true,
			},
		).Get(ctx, nil)

		// execute deletion
		terminated := workflow.Now(ctx)
		account.Terminated = &terminated
	})

	// weekly history refresh
	sel.AddFuture(workflow.NewTimer(ctx, time.Hour*168), func(f workflow.Future) {
		cleanHistory = true
	})

	err := workflow.SetQueryHandler(ctx, "plan", func() (string, error) {
		return account.Plan, nil
	})
	if err != nil {
		return nil, err
	}

	// main loop
	for account.Terminated == nil && !cleanHistory {
		sel.Select(ctx)
	}

	if account.Terminated == nil {
		// transfer state to next RunID
		return nil, workflow.NewContinueAsNewError(ctx, AccountWorkflow, account)
	}

	return account, nil
}
