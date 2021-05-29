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

//UpgradeAccount sends start a proxy workflow from the input and waits for the account workflow's response.
func UpgradeAccount(w http.ResponseWriter, decoder *json.Decoder) {
	// read request to parse out workflow id
	var upgradeRequest *interactive_ui_signal.UpgradeRequest
	err := decoder.Decode(&upgradeRequest)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid request body\"}"))
		return
	}
	if upgradeRequest.Account == "" {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid account\"}"))
		return
	}
	bts, _ := json.Marshal(upgradeRequest)
	upgradeWE, err := temporalClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "interactive-ui-signal",
		},
		proxy.RequestResponse,
		&proxy.Payload{
			TargetId:   upgradeRequest.Account,
			SignalName: interactive_ui_signal.UpgradeAccountSignal,
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
		w.Write([]byte("{\"error\": \"unable to get UpgradeResponse\"}"))
		return
	}
	if result.Error != "" {
		log.Println("ERR: UpgradeRequest failed", result.Error)
		w.Write([]byte("{\"error\": \"UpgradeResponse failed: " + result.Error + "\"}"))
		return
	}

	log.Println("Proxied upgradeAccount", "WorkflowID", upgradeRequest.Account)
	// send through interactive_ui_signal.UpgradeResponse to client
	w.Write(result.ByteData)
}
