package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateDealRequestBody struct {
	DealID string `json:"id" form:"id"`
	Name   string `json:"name" form:"name"`
}

func CreateNewTeamForDeal(c *gin.Context) {

	var requestBody CreateDealRequestBody
	err := c.Bind(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body" + err.Error()})
		return
	}

	if len(requestBody.DealID) == 0 || len(requestBody.Name) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deal Id or Team name not present"})
	}

	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get user " + err.Error()})
	}

	deal, err := models.GetDealByID(requestBody.DealID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get deal " + err.Error()})
	}

	var team models.Team
	team.ID = data.GetUUIDString("deal")
	team.Capacity = deal.TeamCapacity
	team.Name = requestBody.Name
	team.DealID = deal.ID
	team.Strength = 1
	team.CreatedAt = time.Now()

	err = models.CreateTeam(team)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create team" + err.Error()})
	}

	var member models.TeamMember
	member.ID = data.GetUUIDString("deal")
	member.IsAdmin = true
	member.TeamID = team.ID
	member.UserID = user.ID
	err = models.CreateTeamMember(member)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create team member" + err.Error()})
	}

	c.Redirect(http.StatusFound, "/team/"+team.ID)
}

func GetDealDetails(c *gin.Context) {
	dealID := c.Param("dealID")

	deal, err := models.GetDealByID(dealID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get deal " + err.Error()})
		return
	}

	teams, err := models.GetAllTeamsByDealID(dealID, 3)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get teams of a deal " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "deal-teams-list", gin.H{
		"metadata": deal,
		"teams":    teams,
	})
}

func GetDealSummary(c *gin.Context) {

	dealID := c.Param("dealID")

	deal, err := models.GetDealByID(dealID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get deal " + err.Error()})
		return
	}

	product, err := models.GetProduct(deal.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get deal " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "order-summary", gin.H{
		"deal":    deal,
		"product": product,
	})
}

func GetDealTeamSummary(c *gin.Context) {

	variantID := c.Param("variantID")
	dealID := c.Param("dealID")

	deal, err := models.GetDealByID(dealID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get deal " + err.Error()})
		return
	}

	product, err := models.GetProduct(deal.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get deal " + err.Error()})
		return
	}

	variant, err := models.GetVariant(variantID)
	if err != nil {
		fmt.Printf("err variant: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	c.HTML(http.StatusOK, "deal-team", gin.H{
		"deal":    deal,
		"product": product,
		"variant": variant,
	})
}
