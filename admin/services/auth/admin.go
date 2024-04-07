package auth

import (
	"bytes"
	"errors"
	"fmt"
	"hermes/configs"
	"hermes/models"
	"hermes/utils/data"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/datatypes"
)

func AdminLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-login", gin.H{"title": "Login Admin"})
}

func AdminSignInUpUtil(c *gin.Context) error {
	session := sessions.Default(c)

	user := make(map[string]string)
	if err := c.ShouldBind(&user); err != nil {
		fmt.Println("err:", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}

	if user["email"] == "" || user["password"] == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return fmt.Errorf("no email or passoword found")
	}

	admin_email, err := themis.GetSecret("ADMIN_EMAIL")
	if err != nil {
		fmt.Println(err.Error(), admin_email)
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}

	admin_password, err := themis.GetSecret("ADMIN_PASSWORD")
	if err != nil {
		fmt.Println(err.Error(), admin_password)
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}

	if user["email"] != admin_email || user["password"] != admin_password {
		c.AbortWithStatus(http.StatusBadRequest)
		return fmt.Errorf("email or passoword doesn't match")
	}
	admin := CreateOrGetAdminUser(user["email"])
	session.Set(configs.Userkey, admin.ID)
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return fmt.Errorf("error: %s", err.Error())
	}
	return nil
}

func AdminSignInUp(c *gin.Context) {
	err := AdminSignInUpUtil(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Redirect(http.StatusFound, "/")
}

func AdminV2SignInUp(c *gin.Context) {
	err := AdminSignInUpUtil(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response": "success",
	})
}

type GoogleSignInBody struct {
	Token string `json:"token"`
}

func AdminV2SignInUpGoogle(c *gin.Context) {
	session := sessions.Default(c)
	var tokenJson GoogleSignInBody
	c.BindJSON(&tokenJson)
	email, err := getUserEmail(tokenJson.Token)
	if err != nil {
		fmt.Println("ERR: ", err)
		c.AbortWithError(http.StatusUnauthorized, err)
	}
	if !strings.HasSuffix(email, "@amigo.gg") {
		c.AbortWithError(http.StatusForbidden, errors.New("external User"))
		return
	}
	user := CreateOrGetAdminUser(email)
	session.Set(configs.Userkey, user.ID)
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, gin.H{
		"response": "success",
	})
}

func CreateOrGetAdminUser(email string) (user models.Admin) {
	ok, user := checkIfAdminUserExistsAndReturn(email)
	if !ok {
		user, _ = models.CreateAdmin(email)
	}
	return
}

func checkIfAdminUserExistsAndReturn(email string) (userExists bool, user models.Admin) {
	user, _ = models.GetAdmin(email)
	if user.ID != "" {
		userExists = true
	}
	return
}

func AdminAuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)

	if user == nil {
		session.Set(configs.Userkey, nil)
		session.Clear()
		session.Options(sessions.Options{Path: "/", MaxAge: -1})
		session.Save()

		c.JSON(http.StatusUnauthorized, gin.H{"redirect": "/login"})
		c.Abort()
		return
	}
	c.Request.Header.Set("user", user.(string))
	c.Next()
}

func AdminAPIAuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)
	if user == nil {
		session.Set(configs.Userkey, nil)
		session.Clear()
		session.Options(sessions.Options{Path: "/", MaxAge: -1})
		session.Save()

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	_, err := models.GetAdminbyID(user.(string))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			session.Set(configs.Userkey, nil)
			session.Clear()
			session.Options(sessions.Options{Path: "/", MaxAge: -1})
			session.Save()

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	c.Request.Header.Set("user", user.(string))
	c.Next()
}

func SellerAPIAuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)
	if user == nil {
		session.Set(configs.Userkey, nil)
		session.Clear()
		session.Options(sessions.Options{Path: "/", MaxAge: -1})
		session.Save()

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	_, err := models.GetSellerMemberByID(user.(string))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			session.Set(configs.Userkey, nil)
			session.Clear()
			session.Options(sessions.Options{Path: "/", MaxAge: -1})
			session.Save()

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	c.Request.Header.Set("user", user.(string))
	c.Next()
}

func LoggingMiddleware(c *gin.Context) {
	user := c.Request.Header.Get("user")
	portal := c.Request.Header.Get("portal")
	if portal != configs.AdminHeader {
		portal = configs.SellerHeader
	}
	jsonData, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(jsonData))
	params := map[string][]string(c.Request.URL.Query())
	log := models.AdminSellerLogs{
		ID:            data.GetUUIDStringWithoutPrefix(),
		Portal:        models.PortalType(portal),
		UserID:        user,
		RequestMethod: c.Request.Method,
		Body:          string(jsonData),
		Endpoint:      c.Request.URL.Path,
		CreatedAt:     time.Now(),
	}
	paramsToPush := datatypes.JSONMap{}
	for key, val := range params {
		paramsToPush[key] = val
	}
	log.Params = paramsToPush
	log.Create()
	c.Next()
}

func getAuthorizationToken(authHeader string) (token string, err error) {

	authSplit := strings.Split(authHeader, " ")
	if len(authSplit) != 2 {
		return "", fmt.Errorf("auth Header not correctly formatted")
	}

	if authSplit[0] != "Bearer" {
		return "", fmt.Errorf("bearer not present")
	}

	return authSplit[1], nil
}

// Middleware for API Authentication
func APIAccessTokenMiddleware(c *gin.Context) {
	// Get Secret from Header
	token, err := getAuthorizationToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(401, gin.H{
			"error": "Authentication Error",
		})
		c.Abort()
		return
	}

	// Validate Secret
	if token != os.Getenv("ADMIN_API_TOKEN") {
		c.JSON(403, gin.H{
			"error": "Supplied Token Incorrect",
		})
		c.Abort()
		return
	}

	// Allow Disallow Request
	c.Next()
}
