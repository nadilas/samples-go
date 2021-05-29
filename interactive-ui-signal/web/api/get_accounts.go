package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.temporal.io/api/filter/v1"
	"go.temporal.io/api/workflowservice/v1"
)

//GetAccounts queries temporal for open workflow executions and returns the account workflows' details
func GetAccounts(w http.ResponseWriter) {
	openWorkflows, err := temporalClient.ListOpenWorkflow(context.Background(), &workflowservice.ListOpenWorkflowExecutionsRequest{
		Filters: &workflowservice.ListOpenWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: &filter.WorkflowTypeFilter{Name: "AccountWorkflow"},
		},
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"error\": " + err.Error() + "}"))
		return
	}
	bytes, err := json.Marshal(openWorkflows)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"error\": " + err.Error() + "}"))
		return
	}
	w.Write(bytes)
	log.Println("Found", len(openWorkflows.Executions), "account workflows")
}
