package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	interactive_ui_signal "github.com/temporalio/samples-go/interactive-ui-signal"
	"github.com/temporalio/samples-go/interactive-ui-signal/proxy"
	"go.temporal.io/api/filter/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

//go:embed webapp/build
var wwwRoot embed.FS

var temporalClient client.Client

func main() {
	tc, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	temporalClient = tc
	defer tc.Close()

	r := http.NewServeMux()
	webAppRoot, _ := fs.Sub(wwwRoot, "webapp/build")
	webapp := http.FileServer(http.FS(webAppRoot))
	r.HandleFunc("/", corsWrapper(webapp))
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Serve the webapp over TLS
	log.Println("Serving on http://" + srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func corsWrapper(webapp http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "api/") {
			setupResponse(&w)
			apiHandler(w, r)
		} else {
			// host webapp
			webapp.ServeHTTP(w, r)
		}
	}
}

func setupResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "application/json")
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w)

	// assume preflight?
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		w.WriteHeader(200)
		return
	}

	// GET methods
	if r.Method == http.MethodGet {
		switch r.URL.Path {
		case "/api/accounts":
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
			return
		case "/api/plan":
			if account := r.URL.Query().Get("account"); account == "" {
				w.WriteHeader(400)
				w.Write([]byte("{\"error\": \"invalid account\"}"))
				return
			} else {
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
				w.Write([]byte("{\"plan\": \"" + val + "\"}"))
				return
			}
		default:
			http.NotFound(w, r)
			return
		}
	}

	// POST methods
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"invalid request method\"}"))
		return
	}
	decoder := json.NewDecoder(r.Body)
	switch r.URL.Path {
	case "/api/accounts":
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
	case "/api/account/delete":
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

		// send through interactive_ui_signal.UpgradeResponse to client
		w.Write(result.ByteData)
	case "/api/account/upgrade":
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

		// send through interactive_ui_signal.UpgradeResponse to client
		w.Write(result.ByteData)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
