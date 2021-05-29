package api

import (
	"context"
	"log"
	"net/http"
)

//GetAccountPlan queries a specific account workflow for it's current plan state.
func GetAccountPlan(w http.ResponseWriter, r *http.Request) {
	account := ""

	if account = r.URL.Query().Get("account"); account == "" {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid account\"}"))
		return
	}

	ev, err := temporalClient.QueryWorkflow(context.Background(), account, "", "plan")
	if err != nil {
		log.Println("ERR: ", err.Error())
		w.Write([]byte("{\"error\": \"failed to query workflow: " + err.Error() + "\"}"))
		return
	}
	var val string
	err = ev.Get(&val)
	if err != nil {
		log.Println("ERR: ", err.Error())
		w.Write([]byte("{\"error\": \"failed to query workflow: " + err.Error() + "\"}"))
		return
	}
	log.Println("Queried plan", "WorkflowID", account)
	w.Write([]byte("{\"plan\": \"" + val + "\"}"))
}
