package TemporalShared

import "time"

type ACRWorkflowInput struct {
	CartID string
}

type ReviewWorkflowInput struct {
	OrderID  string
	Internal bool
}

type CouponExpiryWorkflowInput struct {
	CouponID string
	MaxLife  time.Duration
}

type ReOrderWorflowInput struct {
	OrderID   string
	Internal  bool
	SleepTime int64
	ProductID string
}

type InfluencerCommisionWorkflowInput struct {
	OrderID    string
	Amount     int64
	CouponCode string
}

type FulfillableOrderWorkflowInput struct {
	OrderID string
}

type NDRWorkflowInput struct {
}

type RoovoPeCreditWorkflowInput struct {
	OrderID  string
	Internal bool
}

type EmbassdorRecruitmentINput struct {
	UserID string
}

type DealActivityInput struct {
	AmbassdorID        string
	ProductID          string
	ChatMessage        string
	TemplateTutorial   string
	TutorialImageLink  string
	TutorialBodyFiller string
}

type InstagramInsightsFetchWorkflowInput struct {
	InfluencerID string
	InstagramID  string
	Followers    int
}

type YoutubeInsightsFetchWorkflowInput struct {
	InfluencerID string
}
