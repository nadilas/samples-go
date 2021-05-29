package api

import "go.temporal.io/sdk/client"

var temporalClient client.Client

func Init() func() {
	tc, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	temporalClient = tc
	return tc.Close
}
