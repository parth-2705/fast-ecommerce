package TemporalJobs

import (
	"context"
	"fmt"
	"hermes/services/Temporal/TemporalShared"
	"hermes/utils/data"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func DeleteTemporalWorkflow2(workflowID string, reason string) (err error) {
	return TemporalShared.C.TerminateWorkflow(context.Background(), workflowID, "", reason)
}

func createJob(wo client.StartWorkflowOptions, workflowToRun interface{}, wi interface{}) (err error) {

	// TemporalShared.C.CancelWorkflow(context.Background(), wo.ID, "")

	_, err = TemporalShared.C.ExecuteWorkflow(context.Background(), wo, workflowToRun, wi)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return

}

func CreateCouponExpiryWorkflow(couponID string, maxLife time.Duration) (err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    couponID,
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		RetryPolicy:           &rp,
		// WorkflowExecutionTimeout: ,
	}

	err = createJob(wo, "CouponExpiryWorkflow", TemporalShared.CouponExpiryWorkflowInput{
		CouponID: couponID,
		MaxLife:  maxLife,
	})
	return
}

func CreateACRWorkflow(cartID string, workflowID string) (err error) {

	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		RetryPolicy:           &rp,
		// WorkflowExecutionTimeout: ,
	}

	err = createJob(wo, "ACRMessagesWorkflow", TemporalShared.ACRWorkflowInput{
		CartID: cartID,
	})

	return
}

func CreateProductReviewMessageWorkflow(orderID string, internal bool) (err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("Review-%s", orderID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		RetryPolicy:           &rp,
		// WorkflowExecutionTimeout: ,
	}

	err = createJob(wo, "ProductReviewWorkflow", TemporalShared.ReviewWorkflowInput{
		OrderID:  orderID,
		Internal: internal,
	})

	return
}

func CreateReOrderMessageWorkflow(orderID string, internal bool, productID string, sleepTime int64) (err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("ReOrder-%s", orderID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		RetryPolicy:           &rp,
		// WorkflowExecutionTimeout: ,
	}

	err = createJob(wo, "ProductReOrderWorkflow", TemporalShared.ReOrderWorflowInput{
		OrderID:   orderID,
		Internal:  internal,
		SleepTime: sleepTime,
		ProductID: productID,
	})

	return
}

func CreateInfluencerCreditsWorkflow(amount int64, couponCode string, orderID string) (err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("ReOrder-%s", orderID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		RetryPolicy:           &rp,
		// WorkflowExecutionTimeout: ,
	}

	err = createJob(wo, "InfluencerCreditCommissionWorkflow", TemporalShared.InfluencerCommisionWorkflowInput{
		OrderID:    orderID,
		CouponCode: couponCode,
		Amount:     amount,
	})

	return
}

func CreateNDRWorkflow() (err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("NDR-%s", data.GetNDRWorkflowID()),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		RetryPolicy:           &rp,
		CronSchedule:          "30 4 * * *",
	}

	err = createJob(wo, "NDRShipmentWorkflow", TemporalShared.NDRWorkflowInput{})

	return
}

func CreateRoovoPeCreditWorkflow(orderID string, internal bool) (err error) {

	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("RoovoPe-%s", orderID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		RetryPolicy:           &rp,
	}

	err = createJob(wo, "RoovoPeCreditWorkflow", TemporalShared.RoovoPeCreditWorkflowInput{
		OrderID:  orderID,
		Internal: internal,
	})

	return
}

func CreateAmbassdorRecruitmentJob(userID string) (wfID string, err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("AMREC-%s", userID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		RetryPolicy:           &rp,
	}

	err = createJob(wo, "SendAmbassdorRecruitmentMessageIndividualWorkflow", TemporalShared.EmbassdorRecruitmentINput{
		UserID: userID,
	})

	return wo.ID, err
}

func CreateAmbassdorDODJob(ambassdorID string, productID string, chatMessage string, tutorialTemplate string, tutorialImage string, tutorialBodyFiller string) (wfID string, err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("AMDEAL-%s", ambassdorID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		RetryPolicy:           &rp,
	}

	err = createJob(wo, "SendDODToAmbassdorWorkflow", TemporalShared.DealActivityInput{
		AmbassdorID:       ambassdorID,
		ProductID:         productID,
		ChatMessage:       chatMessage,
		TemplateTutorial:  tutorialTemplate,
		TutorialImageLink: tutorialImage,
		TutorialBodyFiller: tutorialBodyFiller,
	})

	return wo.ID, err
}

func CreateInstagramInsightsFetchWorkflow(accountID string, followersCount int, influencerID string) (err error) {

	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("Instagram-%s", influencerID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		RetryPolicy:           &rp,
		CronSchedule:          "30 4 * * *",
	}

	err = createJob(wo, "InstagramInsightsWorkflow", TemporalShared.InstagramInsightsFetchWorkflowInput{
		InstagramID:  accountID,
		Followers:    followersCount,
		InfluencerID: influencerID,
	})

	if err != nil {
		return err
	}

	wo1 := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("Influencer-%s", influencerID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		RetryPolicy:           &rp,
	}

	err = createJob(wo1, "InstagramInsightsWorkflow", TemporalShared.InstagramInsightsFetchWorkflowInput{
		InstagramID:  accountID,
		Followers:    followersCount,
		InfluencerID: influencerID,
	})

	return
}

func CreateYoutubeInsightsFetchWorkflow(influencerID string) (err error) {

	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("Youtube-%s", influencerID),
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		RetryPolicy:           &rp,
		CronSchedule:          "30 4 * * *",
	}

	err = createJob(wo, "YouTubeInsightsWorkflow", TemporalShared.YoutubeInsightsFetchWorkflowInput{
		InfluencerID: influencerID,
	})

	if err != nil {
		return err
	}

	return
}

func CreateFulfillableOrderWorkflow(orderID string) (err error) {
	rp := temporal.RetryPolicy{
		MaximumAttempts: 1,
	}
	wo := client.StartWorkflowOptions{
		ID:                    orderID + "fulfillable",
		TaskQueue:             TemporalShared.MessagingQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		RetryPolicy:           &rp,
	}
	err = createJob(wo, "MarkOrderAsFulfillable", TemporalShared.FulfillableOrderWorkflowInput{
		OrderID: orderID,
	})
	return
}
