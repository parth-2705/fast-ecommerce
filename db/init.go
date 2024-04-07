package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

type MongoDBConfig struct {
	URI      string
	Database string
}

type QueryAndCollectionMap struct {
	AggregateQuery bson.A            `json:"aggregateQuery"`
	Collection     *mongo.Collection `json:"collection"`
}

// NewMongoDBClient creates a new MongoDB client
func NewMongoDBClient(config MongoDBConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	return client, nil
}

var ProductCollection *mongo.Collection
var AddressCollection *mongo.Collection
var OrderCollection *mongo.Collection
var PincodeCollection *mongo.Collection
var ReviewCollection *mongo.Collection
var UserCollection *mongo.Collection
var ProfileCollection *mongo.Collection
var WishlistCollection *mongo.Collection
var NotifyCollection *mongo.Collection
var CouponCollection *mongo.Collection
var CategoryCollection *mongo.Collection
var ContactCollection *mongo.Collection
var BrandsCollection *mongo.Collection
var SellerCollection *mongo.Collection
var VariantsCollection *mongo.Collection
var ShippingCollection *mongo.Collection
var PaymentCollection *mongo.Collection
var PayCollection *mongo.Collection
var DealCollection *mongo.Collection
var TeamCollection *mongo.Collection
var TeamMemberCollection *mongo.Collection
var OTPLogCollection *mongo.Collection
var ChatChannelCollection *mongo.Collection
var PageViewCollection *mongo.Collection
var TrackingCollection *mongo.Collection
var UserUserAgentMappingCollection *mongo.Collection
var EventsDumpCollection *mongo.Collection
var CartCollection *mongo.Collection
var QuickRepliesCollection *mongo.Collection
var ShippingCharges *mongo.Collection
var AdminsCollection *mongo.Collection
var BulkAction *mongo.Collection
var SellerMembersCollection *mongo.Collection
var ShiprocketOrderCollection *mongo.Collection
var TransactionCollection *mongo.Collection
var BackwardShipmentCollection *mongo.Collection
var WATemplatesCollection *mongo.Collection
var WalletCollection *mongo.Collection

// Influencer Program
var InfluencerCollection *mongo.Collection
var OAuthStateCollection *mongo.Collection
var InfluencerCampaignApplicationCollection *mongo.Collection

var CommissionCollection *mongo.Collection
var CommissionHistoryCollection *mongo.Collection
var CampaignCollection *mongo.Collection
var PincodeStore *persistence.InMemoryStore
var SessionStore sessions.Store

var StringToMongoCollectionMap map[string]*mongo.Collection

func Connect() {
	// Connect to MongoDB
	mongoDBConfig := MongoDBConfig{
		URI:      os.Getenv("MONGO_URL"),
		Database: os.Getenv("DB_NAME"),
	}

	sessionSecret := os.Getenv("SECRET")

	var err error
	MongoClient, err = NewMongoDBClient(mongoDBConfig)
	if err != nil {
		panic(err)
	}
	ProductCollection = MongoClient.Database(mongoDBConfig.Database).Collection("products")
	CategoryCollection = MongoClient.Database(mongoDBConfig.Database).Collection("categories")
	AddressCollection = MongoClient.Database(mongoDBConfig.Database).Collection("addresses")
	OrderCollection = MongoClient.Database(mongoDBConfig.Database).Collection("orders")
	UserCollection = MongoClient.Database(mongoDBConfig.Database).Collection("users")
	ProfileCollection = MongoClient.Database(mongoDBConfig.Database).Collection("profiles")
	WishlistCollection = MongoClient.Database(mongoDBConfig.Database).Collection("wishlists")
	NotifyCollection = MongoClient.Database(mongoDBConfig.Database).Collection("notify")
	CouponCollection = MongoClient.Database(mongoDBConfig.Database).Collection("coupons")
	PincodeCollection = MongoClient.Database(mongoDBConfig.Database).Collection("pincodes2")
	ReviewCollection = MongoClient.Database(mongoDBConfig.Database).Collection("reviews")
	ContactCollection = MongoClient.Database(mongoDBConfig.Database).Collection("contacts")
	BrandsCollection = MongoClient.Database(mongoDBConfig.Database).Collection("brands")
	SellerCollection = MongoClient.Database(mongoDBConfig.Database).Collection("sellers")
	VariantsCollection = MongoClient.Database(mongoDBConfig.Database).Collection("variants")
	ShippingCollection = MongoClient.Database(mongoDBConfig.Database).Collection("shipping")
	PaymentCollection = MongoClient.Database(mongoDBConfig.Database).Collection("payments")
	PayCollection = MongoClient.Database(mongoDBConfig.Database).Collection("pay")
	DealCollection = MongoClient.Database(mongoDBConfig.Database).Collection("deals")
	TeamCollection = MongoClient.Database(mongoDBConfig.Database).Collection("teams")
	TeamMemberCollection = MongoClient.Database(mongoDBConfig.Database).Collection("teamMembers")
	OTPLogCollection = MongoClient.Database(mongoDBConfig.Database).Collection("OTPLogs")
	ChatChannelCollection = MongoClient.Database(mongoDBConfig.Database).Collection("chat-channels")
	PageViewCollection = MongoClient.Database(mongoDBConfig.Database).Collection("PageView")
	TrackingCollection = MongoClient.Database(mongoDBConfig.Database).Collection("tracking")
	UserUserAgentMappingCollection = MongoClient.Database(mongoDBConfig.Database).Collection("userUserAgentMapping")
	EventsDumpCollection = MongoClient.Database(mongoDBConfig.Database).Collection("eventsDump")
	CartCollection = MongoClient.Database(mongoDBConfig.Database).Collection("carts")
	QuickRepliesCollection = MongoClient.Database(mongoDBConfig.Database).Collection("quickReplies")
	ShippingCharges = MongoClient.Database(mongoDBConfig.Database).Collection("shippingCharges")
	AdminsCollection = MongoClient.Database(mongoDBConfig.Database).Collection("admins")
	BulkAction = MongoClient.Database(mongoDBConfig.Database).Collection("bulkActions")
	SellerMembersCollection = MongoClient.Database(mongoDBConfig.Database).Collection("sellerMembers")
	ShiprocketOrderCollection = MongoClient.Database(mongoDBConfig.Database).Collection("shiprocketOrders")
	TransactionCollection = MongoClient.Database(mongoDBConfig.Database).Collection("transactions")
	BackwardShipmentCollection = MongoClient.Database(mongoDBConfig.Database).Collection("backwardShipments")
	CommissionCollection = MongoClient.Database(mongoDBConfig.Database).Collection("commissions")
	CommissionHistoryCollection = MongoClient.Database(mongoDBConfig.Database).Collection("commissionHistory")
	WATemplatesCollection = MongoClient.Database(mongoDBConfig.Database).Collection("WATemplates")
	WalletCollection = MongoClient.Database(mongoDBConfig.Database).Collection("wallet")
	InfluencerCollection = MongoClient.Database(mongoDBConfig.Database).Collection("influencer")
	OAuthStateCollection = MongoClient.Database(mongoDBConfig.Database).Collection("oauth")
	CampaignCollection = MongoClient.Database(mongoDBConfig.Database).Collection("campaigns")
	InfluencerCampaignApplicationCollection = MongoClient.Database(mongoDBConfig.Database).Collection("influApplication")

	// Create the capped collection options
	capped := true
	sizeInBytes := int64(1048576)
	maxDocuments := int64(1000)
	cappedOptions := &options.CreateCollectionOptions{
		Collation:    nil,
		Capped:       &capped,
		SizeInBytes:  &sizeInBytes,
		MaxDocuments: &maxDocuments,
	}

	// Check if the collection already exists
	collectionExists, err := checkCollectionExists(MongoClient, mongoDBConfig.Database, "oauth")
	if err != nil {
		panic(err)
	}

	fmt.Println("collectionExists:", collectionExists)
	if !collectionExists {
		// Create the capped collection
		err = MongoClient.Database(mongoDBConfig.Database).CreateCollection(context.Background(), "oauth", cappedOptions)
		if err != nil {
			panic(err)
		}
	}

	PincodeStore = persistence.NewInMemoryStore(time.Hour * 24 * 365)

	sessionsCollection := MongoClient.Database(mongoDBConfig.Database).Collection("sessions2")
	SessionStore = mongodriver.NewStore(sessionsCollection, 100000000, true, []byte(sessionSecret))

	StringToMongoCollectionMap = map[string]*mongo.Collection{
		"product":           ProductCollection,
		"brand":             BrandsCollection,
		"category":          CategoryCollection,
		"user":              UserCollection,
		"order":             OrderCollection,
		"profile":           ProfileCollection,
		"variant":           VariantsCollection,
		"payment":           PaymentCollection,
		"review":            ReviewCollection,
		"address":           AddressCollection,
		"deal":              DealCollection,
		"contact":           ContactCollection,
		"pageView":          PageViewCollection,
		"logs":              OTPLogCollection,
		"cart":              CartCollection,
		"admin":             AdminsCollection,
		"bulkAction":        BulkAction,
		"shipping":          ShippingCollection,
		"transaction":       TransactionCollection,
		"backwardShipments": BackwardShipmentCollection,
		"pay":               PayCollection,
		"campaign":          CampaignCollection,
		"influencer":        InfluencerCollection,
	}
}

func checkCollectionExists(client *mongo.Client, dbName, collectionName string) (bool, error) {
	collections, err := client.Database(dbName).ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return false, err
	}

	for _, name := range collections {
		if name == collectionName {
			return true, nil
		}
	}

	return false, nil
}
