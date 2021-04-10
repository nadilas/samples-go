package main

import (
	"log"

	"github.com/temporalio/samples-go/interactive-ui-signal/proxy"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/interactive-ui-signal"
)

func main() {
	// The client and worker are heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "interactive-ui-signal", worker.Options{})

	w.RegisterWorkflow(proxy.RequestResponse)
	w.RegisterWorkflow(interactive_ui_signal.AccountWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
