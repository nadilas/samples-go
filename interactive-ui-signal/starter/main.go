package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	interactive_ui_signal "github.com/temporalio/samples-go/interactive-ui-signal"
	"github.com/temporalio/samples-go/interactive-ui-signal/proxy"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "account_1",
		TaskQueue: "interactive-ui-signal",
	}

	account := &interactive_ui_signal.Account{
		Name: "account_1",
	}
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, interactive_ui_signal.AccountWorkflow, account)
	failOnErr(err, "Unable to execute workflow")
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	log.Println("Simulating client interactions with 5 second delays...")

	// fails, does not own account_1
	time.Sleep(time.Second * 5)
	upgradeAccount(c, "account_2")

	// succeeds from owner: account_1
	time.Sleep(time.Second * 5)
	upgradeAccount(c, "account_1")

	// succeeds from owner: account_1
	time.Sleep(time.Second * 5)
	deleteAccount(c, "account_1")

	err = we.Get(context.Background(), &account)
	failOnErr(err, "Unable to get workflow result")
	log.Printf("Account: %s created at: %v terminated at: %v", account.Name, account.Created, account.Terminated)
}

func deleteAccount(c client.Client, actor string) {
	deleteRequest := interactive_ui_signal.DeleteAccountRequest{
		Actor: actor,
	}
	bts, _ := json.Marshal(deleteRequest)
	deleteWE, err := c.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "interactive-ui-signal",
		},
		proxy.RequestResponse,
		&proxy.Payload{
			TargetId:   "account_1",
			SignalName: interactive_ui_signal.DeleteAccountSignal,
			PayloadData: proxy.PayloadData{
				ByteData: bts,
			},
		},
	)
	failOnErr(err, "Unable to run DeleteRequest")

	// response handling
	var result *proxy.Result
	err = deleteWE.Get(context.Background(), &result)
	failOnErr(err, "Unable to get UpgradeResponse")
	if result.Error != "" {
		log.Println("ERR: DeleteRequest failed", result.Error)
		return
	}

	if !result.Success {
		log.Println("ERR: DeleteRequest was not successful")
		return
	}
	log.Println("DeleteRequest was successful")
}

func upgradeAccount(c client.Client, actor string) {
	// send request
	upgradeRequest := &interactive_ui_signal.UpgradeRequest{
		To:    "premium",
		Actor: actor,
	}
	bts, _ := json.Marshal(upgradeRequest)
	upgradeWE, err := c.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "interactive-ui-signal",
		},
		proxy.RequestResponse,
		&proxy.Payload{
			TargetId:   "account_1",
			SignalName: interactive_ui_signal.UpgradeAccountSignal,
			PayloadData: proxy.PayloadData{
				// send either raw bytes
				ByteData: bts,
				// or send proto.Message
				Data: MustMarshalAny(timestamppb.Now()),
			},
		},
	)
	failOnErr(err, "Unable to run UpgradeRequest")

	// response handling
	var result *proxy.Result
	err = upgradeWE.Get(context.Background(), &result)
	failOnErr(err, "Unable to get UpgradeResponse")
	if result.Error != "" {
		log.Println("ERR: UpgradeRequest failed", result.Error)
		return
	}

	if !result.Success {
		log.Println("ERR: UpgradeRequest was not successful")
		return
	}
	// unmarshal raw bytes to response object
	var upgradeResponse *interactive_ui_signal.UpgradeResponse
	err = json.Unmarshal(result.ByteData, &upgradeResponse)
	failOnErr(err, "Failed to unmarshal UpgradeResponse")
	// or unmarshal response from proto.Message
	updateTs := &timestamp.Timestamp{}
	err = anypb.UnmarshalTo(result.Data, updateTs, proto.UnmarshalOptions{})
	failOnErr(err, "Unable to unmarshal updateTimestamp")
	log.Printf("Account upgrade to %s is valid from %v. Update timestamp=%v\n", upgradeRequest.To, upgradeResponse.ValidFrom, updateTs)
}

func failOnErr(err error, s string) {
	if err != nil {
		log.Fatalln(s, err)
	}
}

func MustMarshalAny(msg proto.Message) *anypb.Any {
	any, err := anypb.New(msg)
	if err != nil {
		log.Fatalln("Unable to marshal proto payload", err)
	}
	return any
}
