package TemporalJobs

import (
	"context"
	"hermes/services/Temporal/TemporalShared"
)

func GetResultOfWorflow(workflowID string, runID string) (result interface{}, err error) {
	err = TemporalShared.C.GetWorkflow(context.Background(), workflowID, runID).Get(context.Background(), &result)
	return
}
