package TemporalShared

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

var C client.Client
var err error

func Initialize() (err error) {
	// Create the client object just once per process
	C, err = client.Dial(client.Options{
		HostPort:  os.Getenv("TEMPORAL_SERVER"),
		Namespace: "roovo",
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}
