package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/temporalio/samples-go/interactive-ui-signal/web/api"
)

//go:embed webapp/build
var wwwRoot embed.FS

func main() {
	apiShutdown := api.Init()
	defer apiShutdown()

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
	log.Println("Serving on http://" + srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func corsWrapper(webapp http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "api/") {
			// route to the api handler
			setupResponseHeaders(&w)
			apiHandler(w, r)
		} else {
			// host webapp
			webapp.ServeHTTP(w, r)
		}
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		// assume preflight?
		w.WriteHeader(200)
		return
	}

	// GET methods
	if r.Method == http.MethodGet {
		switch r.URL.Path {
		case "/api/accounts":
			api.GetAccounts(w)
			return
		case "/api/plan":
			api.GetAccountPlan(w, r)
			return
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
		api.CreateAccount(w, decoder)
	case "/api/account/delete":
		api.DeleteAccount(w, decoder)
	case "/api/account/upgrade":
		api.UpgradeAccount(w, decoder)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// helpers

func setupResponseHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "application/json")
}
