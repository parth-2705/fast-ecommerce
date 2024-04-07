package Temporal

import (
	"fmt"
	"hermes/services/Temporal/TemporalShared"

	"go.temporal.io/sdk/worker"
)

var workflowsToRegister []interface{} = []interface{}{ACRMessagesWorkflow, ProductReviewWorkflow, ProductReOrderWorkflow, NDRShipmentWorkflow, RoovoPeCreditWorkflow, SendAmbassdorRecruitmentMessageIndividualWorkflow, InstagramInsightsWorkflow, CouponExpiryWorkflow, SendDODToAmbassdorWorkflow, YouTubeInsightsWorkflow, MarkOrderAsFulfillable, InfluencerCreditCommissionWorkflow}
var actvitiesToRegister []interface{} = []interface{}{SendACREmail, SendReviewMessage, SendReOrderMessage, NDRShipmentActivity, RoovoPeCreditActivity, InstagramInsightsActivity, ExpireCoupon, SendAmbassdorDeal, YouTubeInsightsActivity, MarkOrderFulfillableActivity, CreditCommissionToInfluencer}

var err error

func Worker() error {

	// This worker hosts both Workflow and Activity functions
	w := worker.New(TemporalShared.C, TemporalShared.MessagingQueue, worker.Options{
		Identity: "Roovo-Worker",
	})

	// Register Workflows
	for i := range workflowsToRegister {
		w.RegisterWorkflow(workflowsToRegister[i])
	}

	// Register Activities
	for i := range actvitiesToRegister {
		w.RegisterActivity(actvitiesToRegister[i])
	}

	// Start listening to the Task Queue
	err = w.Start()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}
