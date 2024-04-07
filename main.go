package main

import (
	"encoding/gob"
	"fmt"
	"hermes/admin"
	GCS "hermes/admin/services/gcs"
	"hermes/configs/Mysql"
	"hermes/configs/Redis"
	"hermes/configs/Sentry"
	"hermes/controllers"
	"hermes/db"
	"hermes/middleware"
	"hermes/models"
	"hermes/routes"
	"hermes/scripts"
	"hermes/search"
	"hermes/services/Temporal"
	"hermes/services/Temporal/TemporalJobs"
	"hermes/services/Temporal/TemporalShared"
	"hermes/utils/amplitude"
	"hermes/utils/chat"
	"hermes/utils/data"
	"hermes/utils/messaging"
	"hermes/utils/payments"
	"hermes/utils/tmpl"
	"hermes/utils/whatsapp"

	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tryamigo/themis"
)

func main() {

	// declare a global http client for the app that can be used for all http requests
	http.DefaultClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	godotenv.Load("configs/.env")
	db.Connect()

	modelsToIndex := []models.IndexCreateable{models.Address{}, models.Brand{}, models.Cart{}, models.Category{}, models.Order{}, models.OTPLog{}, models.PageView{}, models.Product{}, models.QuickReply{}, models.Seller{}, models.Shipping{}, models.ShippingCharges{}, models.ShipRocketTracking{}, models.User{}, models.WishlistObject{}, models.BulkAction{}, models.Coupon{}, models.Review{}, models.WATemplate{}, models.Pincode{}, models.Profile{}, models.Wallet{}}
	for _, model := range modelsToIndex {
		err := model.CreateIndexes()
		if err != nil {
			panic(err)
		}
	}

	err := Redis.Connect()
	if err != nil {
		panic(err)
	}

	err = Mysql.Init()
	if err != nil {
		panic(err)
	}

	err = models.MysqlMigrations()
	if err != nil {
		panic(err)
	}

	err = Sentry.Init()
	if err != nil {
		panic(err)
	}

	isDebug := os.Getenv("IS_DEBUG") == "true"

	err = themis.Init(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), "roovo")
	if err != nil {
		if !isDebug {
			panic(err)
		} else {
			fmt.Println(fmt.Errorf("vault init error: %s", err.Error()))
		}
	}

	err = GCS.Init()
	if err != nil {
		if !isDebug {
			panic(err)
		} else {
			fmt.Println(fmt.Errorf("error: %s", err.Error()))
		}
	}

	err = models.SocialInit()
	if err != nil {
		if !isDebug {
			panic(err)
		} else {
			fmt.Println(fmt.Errorf("error: %s", err.Error()))
		}
	}

	search.LoadSecrets()

	err = TemporalShared.Initialize()
	if err != nil {
		panic(err)
	}

	err = Temporal.Worker()
	if err != nil {
		panic(err)
	}

	whatsapp.Init()
	err = TemporalJobs.CreateNDRWorkflow()
	if err != nil {
		panic(err)
	}

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "scripts" {
		scripts.Run(args)
		return
	}


	// ###########################################################################
	// #################    CLIENT INITIALIZATION          #######################
	// ###########################################################################

	go Init()
	// ###########################################################################
	// #################    ADMIN INITIALIZATION        ##########################
	// ###########################################################################

	admin.Init()

}

func Init() {
	gob.Register(controllers.MobileNumber{})
	gob.Register(models.Product{})
	gob.Register(data.UTMParams{})

	router := gin.Default()
	// db.PincodesInit()
	messaging.Init()
	payments.Init()
	// configs.InitializeFlagsmith2()
	chat.Init()
	chat.InitRoovoSupportUser()

	if os.Getenv("ENVIRONMENT") == "prod" {
		amplitude.Init()
	}

	router.Use(sessions.Sessions("hermes", db.SessionStore))
	router.Use(middleware.AttachErrorAlertMiddleware())
	router.Use(middleware.SessionManagementMiddleware())
	router.Use(middleware.UrlHitsLogging)
	router.SetFuncMap(tmpl.GetFuncMap())

	router.Static("/static/styles", "./static/styles")
	router.Static("/static/js", "./static/js")
	router.Static("static/assets", "./static/assets")
	router.Static("/static/fonts", "./static/fonts")
	router.Static("/static/media", "./static/media")
	router.Static("/static/icons", "./static/icons")
	router.LoadHTMLGlob("static/templates/*.html")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	public := router.Group("/")
	webhooks := router.Group("/")

	private := router.Group("/")
	private.Use(middleware.AuthRequired)

	// // API Routes and the Auth Middleware
	// api := router.Group("/api")
	// api.Use(middleware.APIAuthentication)

	routes.PublicRoutes(public)
	routes.PrivateRoutes(private)
	routes.WebhooksRoutes(webhooks)
	// routes.AuthRoutes(router)
	fmt.Printf("port: %v\n", port)
	routerStartError := router.Run(fmt.Sprintf("0.0.0.0:%s", port))
	if routerStartError != nil {
		panic(routerStartError)
	}
	fmt.Println("Listening and serving on port: ", port)
}
