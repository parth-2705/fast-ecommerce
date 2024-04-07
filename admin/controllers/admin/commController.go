package controllers

import (
	"fmt"
	GCS "hermes/admin/services/gcs"
	"hermes/models"
	"hermes/services/Temporal/TemporalJobs"
	"hermes/utils/rw"
	"hermes/utils/whatsapp"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func getTruncatedList[T any](superList []T, newLength int) []T {

	r := rand.New(rand.NewSource(time.Now().Unix()))

	totalLength := len(superList)
	perm := r.Perm(totalLength)
	var iteratable int = totalLength
	if newLength < totalLength {
		iteratable = newLength
	}

	newSlice := make([]T, 0)

	for i := 0; i < iteratable; i++ {
		newSlice = append(newSlice, superList[perm[i]])
	}

	return newSlice
}

type audience struct {
	Internal bool  `json:"internal"`
	Group    []int `json:"group"`
}

type dealInfo struct {
	ChatMessage        string `json:"chatMessage"`
	ProductID          string `json:"productID"`
	CreativeLink       string `json:"creativeLink"`
	TutorialTemplate   string `json:"tutorialTemplate"`
	TutorialImageLink  string `json:"tutorialImageLink"`
	TutorialBodyFiller string `json:"tutorialBodyFiller"`
}

func getAmbassadorsToSendDealTo(audience audience) ([]models.Ambassdor, error) {
	if audience.Internal {
		return models.GetAllInternalAmbassdors()
	}

	ambassadors := make([]models.Ambassdor, 0)
	groupSet := make(map[int]struct{})
	for _, group := range audience.Group {

		// Only Select a group once
		if _, ok := groupSet[group]; ok {
			continue
		}

		groupSet[group] = struct{}{}

		ambassadorsInGroup, err := models.GetAmbassadorsByTestGroup(group)
		if err != nil {
			return nil, err
		}
		ambassadors = append(ambassadors, ambassadorsInGroup...)
	}

	return ambassadors, nil
}

func SendDealToAmbassdors(c *gin.Context) {

	type reqStruct struct {
		Audience audience `json:"audience"`
		Deal     dealInfo `json:"deal"`
	}

	var reqBody reqStruct
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	// ambassdors, err := getAmbassadorsToSendDealTo(reqBody.Audience)

	_, err = models.GetProduct(reqBody.Deal.ProductID)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	// Currently For Internal Users
	ambassdors, err := getAmbassadorsToSendDealTo(reqBody.Audience)
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	// // Pick Random Ambassdors
	// messagesToSendCount := 100

	// ambassdors = getTruncatedList(ambassdors, messagesToSendCount)

	// Create Workflow that creates Photo to send to User
	total := len(ambassdors)
	success := 0
	failed := 0

	mp := make(map[string]IndRes)

	type wfIDUserIDCOmbo struct {
		WFID   string
		USerID string
	}

	list := make([]wfIDUserIDCOmbo, 0)

	// Download Base Deal Creative
	err = GCS.DownloadImageFromBucket(reqBody.Deal.CreativeLink, "dealCreative")
	if err != nil {
		fmt.Printf("DCS Download err: %v\n", err)
		rw.JSONErrorResponse(c, 400, err)
		return
	}
	defer os.Remove("dealCreative")

	for _, ambassdor := range ambassdors {

		//Create a Temporal Job to send Deal to Ambassdor
		wfID, err := TemporalJobs.CreateAmbassdorDODJob(ambassdor.ID, reqBody.Deal.ProductID, reqBody.Deal.ChatMessage, reqBody.Deal.TutorialTemplate, reqBody.Deal.TutorialImageLink, reqBody.Deal.TutorialBodyFiller)
		if err != nil {
			mp[ambassdor.ID] = IndRes{
				WFID:   wfID,
				Status: false,
				Reason: err.Error(),
			}

			failed++
			continue
		}

		list = append(list, wfIDUserIDCOmbo{
			WFID:   wfID,
			USerID: ambassdor.ID,
		})

	}

	// Get Results for All Workflows
	for _, exec := range list {
		_, err = TemporalJobs.GetResultOfWorflow(exec.WFID, "")
		if err != nil {
			failed++
			mp[exec.USerID] = IndRes{
				WFID:   exec.WFID,
				Status: false,
				Reason: err.Error(),
			}
			continue
		}

		mp[exec.USerID] = IndRes{
			WFID:   exec.WFID,
			Status: true,
			Reason: "",
		}
	}

	response := responseStruct{
		Total:     total,
		Success:   success,
		Failed:    failed,
		AllResuts: mp,
	}

	c.JSON(200, response)
}

type IndRes struct {
	WFID   string `json:"id"`
	Status bool   `json:"status"`
	Reason string `json:"reasons"`
}

type responseStruct struct {
	Total     int               `json:"total"`
	Success   int               `json:"success"`
	Failed    int               `json:"failed"`
	AllResuts map[string]IndRes `json:"results"`
}

func SendAmbassdorRecruitment(c *gin.Context) {

	var commaSepNumbers []string

	err := c.BindJSON(&commaSepNumbers)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	ambassdorsToBe, err := models.GetUsersByPhoneNumbers(commaSepNumbers) // Get Users to Send Communication to from Request Body
	// ambassdorsToBe, err := models.GetAllAmbassdorsTobe()
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	total := len(ambassdorsToBe)
	success := 0
	failed := 0

	mp := make(map[string]IndRes)

	type wfIDUserIDCOmbo struct {
		WFID   string
		USerID string
	}

	list := make([]wfIDUserIDCOmbo, 0)

	for _, ambassdorToBe := range ambassdorsToBe {

		wfID, err := TemporalJobs.CreateAmbassdorRecruitmentJob(ambassdorToBe.ID)
		if err != nil {
			mp[ambassdorToBe.ID] = IndRes{
				WFID:   wfID,
				Status: false,
				Reason: err.Error(),
			}

			failed++
			continue
		}

		list = append(list, wfIDUserIDCOmbo{
			WFID:   wfID,
			USerID: ambassdorToBe.ID,
		})

	}

	for _, exec := range list {
		_, err = TemporalJobs.GetResultOfWorflow(exec.WFID, "")
		if err != nil {
			failed++
			mp[exec.USerID] = IndRes{
				WFID:   exec.WFID,
				Status: false,
				Reason: err.Error(),
			}
			continue
		}

		mp[exec.USerID] = IndRes{
			WFID:   exec.WFID,
			Status: true,
			Reason: "",
		}
	}

	response := responseStruct{
		Total:     total,
		Success:   success,
		Failed:    failed,
		AllResuts: mp,
	}

	c.JSON(200, response)
}

type InfluencerStruct struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func SendInfluencerRecruitment(c *gin.Context) {
	var influencerLeads []InfluencerStruct

	err := c.BindJSON(&influencerLeads)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(influencerLeads) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No influencer leads found"})
		return
	}

	total := len(influencerLeads)

	for _, lead := range influencerLeads {
		err = whatsapp.SendInfluencerOnboardingTemplate(lead.Name, lead.Phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}
