package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTeamDetails(c *gin.Context) {
	teamID := c.Param("teamID")

	team, err := models.GetTeamByID(teamID)
	if err != nil {
		fmt.Println("team not found" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"team not found": err.Error()})
		return
	}

	deal, err := models.GetDealByID(team.DealID)
	if err != nil {
		fmt.Println("deal not found" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"deal not found": err.Error()})
		return
	}

	teamMembers, err := models.GetTeamMembersByTeamID(teamID, -1)
	if err != nil {
		fmt.Println("team members not found" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var isUserAMember bool
	user, err := getUserObjectFromSession(c)
	if err != nil {
		isUserAMember = false
	} else {
		isUserAMember = models.IsUserTeamMember(user, teamMembers)
	}

	c.HTML(http.StatusOK, "teamPage", gin.H{
		"deal":          deal,
		"team":          team,
		"teamMembers":   teamMembers,
		"isUserAMember": isUserAMember,
	})
}

func JoinTeam(c *gin.Context) {
	user, _ := getUserObjectFromSession(c)

	teamID := c.Query("teamID")

	var member models.TeamMember
	member.ID = data.GetUUIDString("team")
	member.TeamID = teamID
	member.IsAdmin = false
	member.UserID = user.ID

	err := models.CreateTeamMember(member)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, member)
}
