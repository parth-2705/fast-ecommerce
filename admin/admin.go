package admin

import (
	"fmt"
	controllers "hermes/admin/controllers/admin"
	"hermes/admin/controllers/common"
	"hermes/admin/services/auth"
	GCS "hermes/admin/services/gcs"
	"hermes/configs"
	"hermes/middleware"
	"hermes/seller"
	"hermes/services/shiprocket"
	"hermes/utils/tmpl"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func Init() {

	// ###########################################################################
	// #################    ADMIN INITIALIZATION        ##########################
	// ###########################################################################
	go AdminInit()

	// ###########################################################################
	// #################    SELLER INITIALIZATION        #########################
	// ###########################################################################
	seller.Init()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := os.Getenv("ADMIN_FRONTEND_URL")

		if len(origin) == 0 {
			origin = "https://admin.roovo.in"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, portal")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func AdminInit() {

	admin := gin.Default()

	admin.SetFuncMap(tmpl.GetFuncMap())

	admin.Use(sessions.Sessions("admin3", cookie.NewStore(configs.Secret)))
	admin.Use(middleware.AttachErrorAlertMiddleware())
	admin.Use(CORSMiddleware())
	public := admin.Group("")
	private := admin.Group("")
	publicV2 := admin.Group("v2")
	privateV2 := admin.Group("v2")
	api := admin.Group("api")

	private.Use(auth.AdminAuthMiddleware)
	private.Use(auth.LoggingMiddleware)
	privateV2.Use(auth.AdminAPIAuthMiddleware)
	privateV2.Use(auth.LoggingMiddleware)
	api.Use(auth.APIAccessTokenMiddleware)

	// New Admin UI on next
	AdminV2RouterGroup(privateV2)
	AdminV2PublicRouterGroup(publicV2)
	AdminV2APIRouterGroup(api)

	admin.Static("/admin/static/styles", "./admin/static/styles")
	admin.Static("admin/static/assets", "./admin/static/assets")
	admin.Static("/admin/static/fonts", "./admin/static/fonts")
	admin.Static("/admin/static/js", "./admin/static/js")
	admin.LoadHTMLGlob("admin/static/templates/*/**")

	// Product UI
	// private.GET("/", controllers.AllProductsPage2)
	// private.GET("/products", controllers.AllProductsPage)
	// private.GET("/product/new", controllers.NewProductsPage)
	// private.GET("/product/:productId", controllers.ProductEditPage)

	// Product UI
	private.GET("/", controllers.AllOrdersPage)
	private.GET("/products", controllers.AllProductsPage2)
	private.GET("/product/new", controllers.NewProductsPage2)
	private.GET("/product/:productId", controllers.ProductEditPage2)
	private.GET("/product/get-page", controllers.GetProductPage)

	// Product UI
	private.GET("/products2", controllers.AllProductsPage2)
	private.GET("/product2/new", controllers.NewProductsPage2)
	private.GET("/product2/:productId", controllers.ProductEditPage2)

	// Product Model CRUD
	private.POST("/product", controllers.AddProduct)
	private.PUT("/product", controllers.UpdateProduct)      // with id as query param
	private.DELETE("/product", controllers.DeleteProduct)   // with id as query param
	private.PATCH("/product", controllers.DuplicateProduct) // with id as query param

	// Category UI
	private.GET("/categories", controllers.AllCategoriesPage)
	private.GET("/category/new", controllers.NewCategoryPage)
	private.GET("/category/:categoryId", controllers.CategoryEditPage)

	// Category Model CRUD
	private.POST("/category", controllers.CreateCategory)
	private.PUT("/category", controllers.UpdateCategory)
	private.DELETE("/category", controllers.DeleteCategory)

	// Customers UI
	private.GET("/customers", controllers.AllCustomersPage)

	//Variants
	private.GET("/variants", controllers.GetVariants)
	private.DELETE("/batch-delete-variants", controllers.BatchDeleteVariants)
	private.POST("/batch-add-variants", controllers.BatchAddVariants)

	// Brands UI
	private.GET("/brands", controllers.AllBrandsPage)
	private.GET("/brand/new", controllers.NewBrandPage)
	private.GET("/brand/:brandId", controllers.BrandEditPage)

	//Brand Model CRUD
	private.POST("/brand", controllers.CreateBrand)
	private.PUT("/brand", controllers.UpdateBrand)
	private.DELETE("/brand", controllers.DeleteBrand)

	// Deals UI
	private.GET("/deals", controllers.AllDealsPage)
	private.GET("/deal/new", controllers.NewDealPage)
	private.GET("/deal/:dealID", controllers.DealEditPage)
	private.POST("/deal", controllers.DealCreate)
	private.PUT("/deal", controllers.DealEdit)
	private.DELETE("/deal", controllers.SetDealInactive)

	// Utilities
	private.POST("/upload-image", GCS.UploadProfileImage)
	private.POST("/upload-media", GCS.UploadMedia)
	private.POST("/upload-csv", GCS.UploadBulkActionCSV)
	private.GET("/addresses", controllers.GetAddressesForUser)

	//Orders UI
	private.GET("/orders", controllers.AllOrdersPage)
	private.GET("/order/new", controllers.NewOrderPage)
	private.GET("/order/:orderId", controllers.OrderViewPage)

	// Order Model CRUD
	// private.POST("/order", controllers.AddOrder)
	private.PUT("/order", controllers.UpdateOrder)
	private.DELETE("/order", controllers.DeleteOrder)

	// Auth
	public.GET("/login", auth.AdminLogin)     // /login
	public.POST("/login", auth.AdminSignInUp) // /login form
	private.GET("/sign-out", auth.SignOut)

	// Chat UI
	private.GET("/chats/:chatID", controllers.AllChatsPage)

	// Invoice UI
	private.GET("/invoice", common.GenerateInvoice)
	private.GET("/seller/invoice", common.GenerateSellerInvoice)

	//Shiprocket
	private.GET("/couriers", shiprocket.GetAllCouriers)
	private.GET("/ship-now", shiprocket.CreateAdminOrderOnShipRocket)
	private.GET("/generate-pick-up", shiprocket.CreatePickupOnShipRocket)
	private.GET("/generate-manifest", shiprocket.GenerateOrderManifest)
	private.GET("/add-pickup-address", shiprocket.AddShiprocketPickUpAddress)

	private.GET("/reassign", shiprocket.ReassignShipment)
	private.POST("/create-return", shiprocket.CreateReturnForOrderOnShiprocket)
	private.POST("/missing-order", shiprocket.CreateShipmentForMissingItems)
	private.POST("/replacement-order", shiprocket.CreateShipmentForReplacementItems)
	private.POST("/exchange-order", shiprocket.CreateExchangeForOrderOnShiprocket)

	private.POST("/cancel-return-order", shiprocket.CancelReturnOrder)

	private.GET("/shipping-charges", shiprocket.GetShippingCharges)
	private.GET("/mark-dispatched", shiprocket.MarkDispatch)
	private.GET("/unmark-dispatched", shiprocket.UnmarkDispatch)
	private.GET("/mark-processed", shiprocket.MarkProcessed)
	private.GET("/unmark-processed", shiprocket.UnmarkProcessed)

	adminPort := os.Getenv("ADMIN_PORT")
	fmt.Println("adminPort:", adminPort)
	admin.Run(":" + adminPort)
	fmt.Println("Listening and serving on port: 8999")
}

func AdminV2RouterGroup(g *gin.RouterGroup) {

	// Orders List

	g.DELETE("/batch-delete-variants", controllers.BatchDeleteVariants)
	g.POST("/batch-add-variants", controllers.BatchAddVariants)
	g.POST("/batch-replace-variants", controllers.BatchReplaceVariants)

	orders := g.Group("/orders")
	{
		orders.GET("", controllers.GetOrdersList)
		orders.GET("/download-report", controllers.DownloadOrderReport)
		orders.POST("/bulk-ship", controllers.CreateBulkShipment)
	}

	// Order Creation
	order := g.Group("/order")
	{
		order.POST("/new", controllers.CreateNewOrder)
		order.POST("/cancel-order", controllers.CancelOrder)
		order.GET("/summary/:phone", controllers.GetOrdersSummaryForUserByPhone)
		order.POST("/mark-fulfillable", controllers.ConfirmCODPayment)
	}

	// Variant
	variant := g.Group("/variant")
	{
		variant.GET("", controllers.GetProductVariant)
		variant.GET("/all", controllers.GetAllProductVariants)
		variant.POST("", controllers.EditVariantDetails)
		variant.POST("/bulk-update", controllers.BulkUpdate)
	}

	// Shipment
	shipment := g.Group("/shipment")
	{
		shipment.POST("mark-as-downloaded", common.MarkShipmentItemAsDownloaded)
	}

	// Products List
	products := g.Group("/products")
	{
		products.GET("", controllers.GetProductList)
		products.GET("/search", controllers.GetProductList2)
	}

	// Product Filtering
	product := g.Group("/product")
	{
		product.GET("", controllers.GetProductByID)
		product.GET("/find", controllers.FilterProductByName)
		product.POST("", controllers.AddProduct)
		product.PUT("", controllers.UpdateProduct)
	}

	//campaigns
	campaign := g.Group("/campaigns")
	{
		campaign.GET("", controllers.GetCampaignByID)
		campaign.GET("/all", controllers.GetAllCampaigns)
		campaign.POST("", controllers.CreateCampaign)
		campaign.PUT("", controllers.UpdateCampaign)
	}

	// OTP Logs List
	otpLogs := g.Group("/otp-logs")
	{
		otpLogs.GET("", controllers.GetOTPLogsList)
	}

	// Page Views List
	pageViews := g.Group("/page-views")
	{
		pageViews.GET("", controllers.GetPageViews)
	}

	// Cart
	carts := g.Group("/carts")
	{
		carts.GET("", controllers.GetCartsList)
	}

	// Quick Replies CRUD
	quickReplies := g.Group("/quick-reply")
	{
		quickReplies.GET("", controllers.GetAllQuickReplies)
		quickReplies.POST("", controllers.CreateQuickReply)
		quickReplies.PATCH("", controllers.UpdateQuickReply)
		quickReplies.DELETE("", controllers.DeleteQuickReply)

	}

	categories := g.Group("/categories")
	{
		categories.GET("", controllers.GetCompleteCategoryList)
	}

	category := g.Group("/category")
	{
		category.GET("", controllers.GetCategoryListPage)
		category.GET("/get-page", controllers.GetCategoryPage)
		category.GET("/:id", controllers.GetCategoryByID)
		category.POST("", controllers.CreateCategory)
		category.PUT("", controllers.UpdateCategory)
	}

	users := g.Group("/users")
	{
		users.GET("", controllers.GetUsers)
		users.GET("/get-page", controllers.GetUserPage)
		users.GET("/addresses", controllers.GetAllUserAddressess)
		users.POST("", controllers.CreateUser)
		users.PUT("", controllers.UpdateUser)
	}

	user := g.Group("/user")
	{
		user.GET("/internal", controllers.GetUserInternalStatus)
		user.POST("/toggle-internal-status", controllers.ToggleUserInternalExternalStatus)
	}

	// Auth Status
	authGroup := g.Group("/auth")
	{
		authGroup.GET("/status", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"response": "Logged in"})
		})
		authGroup.GET("/sign-out", auth.SignOut)
		authGroup.GET("/admin", controllers.GetAdminInfo)
	}

	pincodes := g.Group("/pincodes")
	{
		pincodes.GET("/:pincode", controllers.GetDataByPincode)
		pincodes.GET("/states", controllers.GetListOfStates)
	}

	// Sellers
	sellers := g.Group("/sellers")
	{
		sellers.GET("", controllers.GetListOfSellers)
	}

	// Seller CRUD
	seller := g.Group("/seller")
	{
		seller.GET("/:sellerID", controllers.GetSellerByID)
		seller.POST("", controllers.CreateSeller)
		seller.PUT("", controllers.UpdateBrand)
	}

	// Brands
	brands := g.Group("/brands")
	{
		brands.GET("", controllers.GetBrandsList)
	}

	// Brand CRUD
	brand := g.Group("/brand")
	{
		brand.GET("/:brandID", controllers.GetBrandByID)
		brand.POST("", controllers.CreateBrand)
		brand.PUT("", controllers.UpdateBrand)
	}

	transactions := g.Group("/transactions")
	{
		transactions.GET("", controllers.GetAllTransactions)
		transactions.GET("/remittable-details", controllers.GetRemittableDetailsForTimeRange)
	}

	transaction := g.Group("/transaction")
	{
		transaction.GET("", controllers.GetTransactionByID)
		transaction.POST("", controllers.CreateTransaction)
		transaction.PUT("", controllers.UpdateTransaction)
		transaction.GET("/download-report/:id", controllers.DownloadTransactionReport)
	}

	returns := g.Group("/returns")
	{
		returns.GET("", controllers.GetAllReturns)
	}

	// Bulk Actions
	bulkAction := g.Group("/bulk")
	{
		bulkAction.POST("/bulk-insert", controllers.InsertCatalogue)
	}

	// Influencer
	influencer := g.Group("/influencer")
	{
		influencer.GET("/list", controllers.GetInfluencersList)
		influencer.POST("/approve", controllers.ApproveInfluencer)
		influencer.POST("/disapprove", controllers.DisapproveInfluencer)
		influencer.POST("", controllers.UpdateInstagramProfile)
		influencer.POST("/campaign/approve/:id", controllers.ApproveInfluencerCampaignApplication)
	}

	//Comms Endpoint
	comm := g.Group("/comm")
	{
		ambassdor := comm.Group("/ambassdor")
		{
			ambassdor.POST("DOD", controllers.SendDealToAmbassdors)
		}
	}

}

func AdminV2PublicRouterGroup(g *gin.RouterGroup) {
	// Auth Status
	authGroup := g.Group("/auth")
	{
		authGroup.POST("/login", auth.AdminV2SignInUp)
		authGroup.POST("/login/google", auth.AdminV2SignInUpGoogle)
	}
}

func AdminV2APIRouterGroup(g *gin.RouterGroup) {
	// API Router
	health := g.Group("/health")
	{
		health.GET("", controllers.HealthCheck)
	}

	product := g.Group("/product")
	{
		product.GET("", controllers.GetProductByID)
		product.PUT("/review", controllers.AddReviewForProduct)
	}

	order := g.Group("/order")
	{
		order.POST("", controllers.CreateNewOrderFromCart)
		order.POST("/convert-to-prepaid", controllers.ConvertPaymentToPrepaid)
		order.POST("/confirm-cod", controllers.ConfirmCODPayment)
		order.POST("/cancel", controllers.CancelOrder)
	}

	cart := g.Group("/cart")
	{
		cart.POST("/recover", controllers.RecoverCartOnWhatsapp)
		cart.POST("", controllers.CreateNewCart)
	}

	address := g.Group("/address")
	{
		address.POST("/:phone", controllers.CreateDefaultAddress)
		address.GET("/:phone", controllers.GetDefaultAddress)
	}

	user := g.Group("/user")
	{
		user.POST("/stop-market-comm", controllers.StopUserMarketComm)
		user.POST("", controllers.CreateUser)
		user.POST("/referral", controllers.JoinReferralProgram)
		user.POST("/referral/mass", controllers.MakeAmbassdorsFromCSV)
		user.GET("/:phone", controllers.GetUserAPI)
	}

	ambassdor := g.Group("/ambassdor")
	{
		ambassdor.POST("/click", controllers.RecordProductViewForAmbassdor)
	}

	comm := g.Group("/comm")
	{
		comm.POST("/ambassdor", controllers.SendAmbassdorRecruitment)
		comm.POST("/ambassdor/DOD", controllers.SendDealToAmbassdors)
		comm.POST("/influencer/onboarding", controllers.SendInfluencerRecruitment)
	}

	g.POST("/reschedule-shipment", controllers.ReschduleShipment)
	g.POST("/update-shipment-address", controllers.UpdateShipmentAddress)
	g.POST("/cancel-shipment", controllers.CancelShipment)
	g.POST("/fake-shipment-attempt", controllers.ReportShipmentAsFakeAttempt)
}
