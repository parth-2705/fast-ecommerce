package Temporal

import (
	"hermes/services/Temporal/TemporalShared"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	min20  = time.Minute * 20
	hour1  = time.Hour * 1
	hour4  = time.Hour * 4
	hour24 = time.Hour * 24
	day3   = time.Hour * 24 * 3
	day2   = time.Hour * 24 * 2
	day4   = time.Hour * 24 * 4
	day8   = time.Hour * 24 * 8
)

func ACRMessagesWorkflow(ctx workflow.Context, wi TemporalShared.ACRWorkflowInput) error {

	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 50,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	ai := ACRActivityInput{
		CartID: wi.CartID,
	}

	timeElapsed := 0 * time.Second

	ai.MessageSenderID = 0                          // Set the Message Sender
	workflow.Sleep(ctx, hour1-timeElapsed)          // Wait till next Activity Run
	workflow.ExecuteActivity(ctx, SendACREmail, ai) // Run Activity
	timeElapsed += hour1                            // Add the slept time duration to elapsedtime

	ai.MessageSenderID = 1
	workflow.Sleep(ctx, hour4-timeElapsed)
	workflow.ExecuteActivity(ctx, SendACREmail, ai)
	timeElapsed += hour4

	ai.MessageSenderID = 2
	workflow.Sleep(ctx, hour24-timeElapsed)
	workflow.ExecuteActivity(ctx, SendACREmail, ai)
	timeElapsed += hour24

	ai.MessageSenderID = 3
	workflow.Sleep(ctx, day3-timeElapsed)
	workflow.ExecuteActivity(ctx, SendACREmail, ai).Get(ctx, nil)
	timeElapsed += day3

	return nil
}

func CouponExpiryWorkflow(ctx workflow.Context, wi TemporalShared.CouponExpiryWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 50,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	ai := CouponExpiryActivityInput{
		CouponID: wi.CouponID,
	}
	workflow.Sleep(ctx, wi.MaxLife)
	workflow.ExecuteActivity(ctx, ExpireCoupon, ai).Get(ctx, nil)
	return nil
}

func ProductReviewWorkflow(ctx workflow.Context, wi TemporalShared.ReviewWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 50,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	ai := ReviewActivityInput{
		OrderID: wi.OrderID,
	}

	if !wi.Internal {
		workflow.Sleep(ctx, day2)
	}
	err = workflow.ExecuteActivity(ctx, SendReviewMessage, ai).Get(ctx, nil)
	return err
}

func InfluencerCreditCommissionWorkflow(ctx workflow.Context, wi TemporalShared.InfluencerCommisionWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 50,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	ai := InfluencerCommissionActivityInput{
		OrderID:    wi.OrderID,
		Amount:     wi.Amount,
		CouponCode: wi.CouponCode,
	}
	workflow.Sleep(ctx, time.Hour*24*7)
	err = workflow.ExecuteActivity(ctx, CreditCommissionToInfluencer, ai).Get(ctx, nil)
	return err
}

func ProductReOrderWorkflow(ctx workflow.Context, wi TemporalShared.ReOrderWorflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 50,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	ai := ReorderActivityInput{
		OrderID: wi.OrderID,
	}

	if !wi.Internal {
		workflow.Sleep(ctx, time.Duration(wi.SleepTime)*time.Hour*24)
	}
	err = workflow.ExecuteActivity(ctx, SendReOrderMessage, ai).Get(ctx, nil)
	return err
}

func NDRShipmentWorkflow(ctx workflow.Context, wi TemporalShared.NDRWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 60,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	err = workflow.ExecuteActivity(ctx, NDRShipmentActivity, wi).Get(ctx, nil)
	return err
}

func RoovoPeCreditWorkflow(ctx workflow.Context, wi TemporalShared.RoovoPeCreditWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 0,
		InitialInterval: time.Hour * 24,
		MaximumInterval: time.Hour * 24,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 60,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	if !wi.Internal {
		err = workflow.Sleep(ctx, day8)
		if err != nil {
			return err
		}
	}

	// Activity to add 5% of the order value to the referring user wallet
	err = workflow.ExecuteActivity(ctx, RoovoPeCreditActivity, wi).Get(ctx, nil)

	if err != nil {
		return err
	}

	// Activity to add 1% of the order value to the ordering user wallet
	err = workflow.ExecuteActivity(ctx, CashbackCreditActivity, wi).Get(ctx, nil)
	return err
}

func SendAmbassdorRecruitmentMessageIndividualWorkflow(ctx workflow.Context, wi TemporalShared.EmbassdorRecruitmentINput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.LocalActivityOptions{
		StartToCloseTimeout: time.Second * 60,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithLocalActivityOptions(ctx, options)

	err = workflow.ExecuteLocalActivity(ctx, SendAmbassdorMessageActivity, wi).Get(ctx, nil)
	return err
}

func SendDODToAmbassdorWorkflow(ctx workflow.Context, wi TemporalShared.DealActivityInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 60,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	err = workflow.ExecuteActivity(ctx, SendAmbassdorDeal, wi).Get(ctx, nil)
	return err
}

func InstagramInsightsWorkflow(ctx workflow.Context, wi TemporalShared.InstagramInsightsFetchWorkflowInput) error {

	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.LocalActivityOptions{
		StartToCloseTimeout: time.Second * 60,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithLocalActivityOptions(ctx, options)

	err = workflow.ExecuteLocalActivity(ctx, InstagramInsightsActivity, wi).Get(ctx, nil)
	return err
}

func YouTubeInsightsWorkflow(ctx workflow.Context, wi TemporalShared.YoutubeInsightsFetchWorkflowInput) error {

	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.LocalActivityOptions{
		StartToCloseTimeout: time.Second * 60,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithLocalActivityOptions(ctx, options)

	err = workflow.ExecuteLocalActivity(ctx, YouTubeInsightsActivity, wi).Get(ctx, nil)
	return err
}

func MarkOrderAsFulfillable(ctx workflow.Context, wi TemporalShared.FulfillableOrderWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy:         retryPolicy,
	}

	err = workflow.Sleep(ctx, time.Minute)
	if err != nil {
		return err
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	err = workflow.ExecuteActivity(ctx, MarkOrderFulfillableActivity, wi).Get(ctx, nil)
	return err
}
