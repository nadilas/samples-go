package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/temporalio/samples-go/interactive-ui-signal"
	"go.temporal.io/sdk/client"
)

//CreateAccount starts a new perpetual account workflow from the input.
func CreateAccount(w http.ResponseWriter, decoder *json.Decoder) {
	var account *interactive_ui_signal.Account
	err := decoder.Decode(&account)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid request body\"}"))
		return
	}
	if account.Name == "" {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"account name is required\"}"))
	} else {
		workflowOptions := client.StartWorkflowOptions{
			ID:        account.Name,
			TaskQueue: "interactive-ui-signal",
		}
		if account.Plan == "" {
			account.Plan = "trial"
		}
		we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, interactive_ui_signal.AccountWorkflow, account)
		if err != nil {
			log.Println("ERR: ", err.Error())
			w.Write([]byte("{\"error\": \"unable to execute workflow\"}"))
			return
		}
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}
}
