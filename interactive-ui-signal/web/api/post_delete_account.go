package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/temporalio/samples-go/interactive-ui-signal"
	"github.com/temporalio/samples-go/interactive-ui-signal/proxy"
	"go.temporal.io/sdk/client"
)

//DeleteAccount sends start a proxy workflow from the input and waits for the account workflow's response.
func DeleteAccount(w http.ResponseWriter, decoder *json.Decoder) {
	// read request to parse out workflow id
	var deleteRequest *interactive_ui_signal.DeleteAccountRequest
	err := decoder.Decode(&deleteRequest)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid request body\"}"))
		return
	}
	if deleteRequest.Account == "" {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid account\"}"))
		return
	}
	bts, _ := json.Marshal(deleteRequest)
	upgradeWE, err := temporalClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "interactive-ui-signal",
		},
		proxy.RequestResponse,
		&proxy.Payload{
			TargetId:   deleteRequest.Account,
			SignalName: interactive_ui_signal.DeleteAccountSignal,
			PayloadData: proxy.PayloadData{
				// send either raw bytes
				ByteData: bts,
			},
		},
	)
	if err != nil {
		log.Println("ERR: ", err.Error())
		w.Write([]byte("{\"error\": \"unable to signal workflow\"}"))
		return
	}
	var result *proxy.Result
	err = upgradeWE.Get(context.Background(), &result)
	if err != nil {
		log.Println("ERR: ", err.Error())
		w.Write([]byte("{\"error\": \"unable to get DeleteRequest\"}"))
		return
	}
	if result.Error != "" {
		log.Println("ERR: DeleteRequest failed", result.Error)
		w.Write([]byte("{\"error\": \"DeleteRequest failed: " + result.Error + "\"}"))
		return
	}

	log.Println("Proxied deleteAccount", "WorkflowID", deleteRequest.Account)
	// send through interactive_ui_signal.UpgradeResponse to client
	w.Write(result.ByteData)
}
