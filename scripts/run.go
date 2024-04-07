package scripts

import (
	"fmt"
	"hermes/configs/Redis"
	"hermes/services/shiprocket"
	"strconv"
)

func Run(args []string) {
	var err error
	fmt.Println("ARGS", args)
	if args[0] == "scripts" {
		scriptToRun := args[1]
		switch scriptToRun {
		case "addDocumentsToSearchDB":
			err = AddCatalogToIndex()
		case "deleteAllDocumentsFromSearchDB":
			err = DeleteAllDocuments()
		case "removeStatusFromCart":
			err = removeStatusFromCarts()
		case "removeURLFromImages":
			err = removeURLFromImages()
		case "addOrReplaceProduct":
			err = AddOrReplaceProduct(args[2])
		case "encodeProductDescriptions":
			err = encodeDescriptionsForProduct()
		case "addOrdersToOrderLogs":
			err = addOrdersToOrderLogs()
		case "addUserIDToInfluencerModel":
			err = addUserIDToInfluencerModel()
		case "addShiprocketOrdersToOrderLogs":
			err = addShiprocketOrdersToOrderLogs()
		case "populateShippingCharges":
			err = populateShippingCharges()
		case "addShippingsToLogs":
			err = addShippingsToLogs()
		case "addProductsFromJson":
			err = addProductsFromJson()
		case "removeRedisCache":
			err = Redis.DeleteProductCacheByID(args[2])
		case "removeMeiliProduct":
			err = DeleteDocumentFromSearchDB(args[2])
		case "updateEmptyNewCategory":
			err = updateEmptyNewCategory()
		case "makeSellerMembersFromSellers":
			err = makeSellerMembersFromSellers()
		case "populateShiprocketOrders":
			err = shiprocket.PopulateShiprocketOrders()
		case "categoriesImgURLFix":
			err = fixImageURLsInCategories()
		case "updateEANCode":
			err = updateEANCode(args[2], args[3])
		case "fillCartInOrders":
			err = fillCartInOrders()
		case "removeStreetNameFromAddress":
			err = removeStreetNameFromAddress()
		case "updateShippingParent":
			err = updateShippingParent()
		case "createReturn":
			err = createReturn(args[2])
		case "addPaymentMethodsToProducts":
			err = addPaymentMethodsToProducts()
		case "addGSTToOrder":
			err = addGSTToOrder()
		case "defaultAddress":
			err = makeOnlyAddressDefault()
		case "updateAWBCodeForShipping":
			err = updateAWBCodeForShipping()
		case "testtemplate":
			err = templateTest(args[2])
		case "addUserObjectToOrders":
			err = addUserObjectToOrders()
		case "deleteOrderIfUserNotFound":
			err = deleteOrderIfUserNotFound()
		case "addUserEmailByPhone":
			err = addUserEmailByPhone()
		case "updateShiprocketOrder":
			temp, err := strconv.Atoi(args[2])
			if err == nil {
				err = shiprocket.UpdateShiprocketOrder(temp)
			}
		case "fixAdminUsersInCarts":
			err = fixAdminUsersInCarts()
		case "unremitShipmentsFromTransaction":
			err = unremitShipmentsFromTransaction(args[2])
		case "remitShipmentsFromTransaction":
			err = remitShipmentsFromTransaction(args[2])
		case "removeMultipleRatings":
			err = removeMultipleRatings()
		case "addRepeatUserKey":
			err = addRepeatUserKeyToUsers()
		case "addCommissionToOrders":
			err = addCommissionToOrders()
		case "exportCartsToMySql":
			err = exportCartsToMySql()
		case "exportBackwardShipmentsToMySql":
			err = exportBackwardShipmentsToMySql()
		case "addRTODeliveredTimestamp":
			err = shiprocket.AddRTODeliveredTimestamp()
		case "addWalletToUser":
			err = addWalletToUser()
		case "deleteProductFromRedisBySellerID":
			err = deleteProductFromRedisBySellerID(args[2])
		case "CancelAllOrdersOfASeller":
			// takes seller ID and reason for cancellation
			err = CancelAllOrdersOfASeller(args[2], args[3])
		case "CreateDuplicateOrderBySeller":
			err = CreateDuplicateOrderBySeller(args[2], args[3])
		case "AddDeliveryTimeToProducts":
			err = AddDeliveryTimeToProducts()
		case "createCoupon":
			err = createCoupon()
		case "migrateCoupons":
			err = MigrateOldCouponStructtoNew()
		case "migrateAmbassdors":
			err = makeAmbassdorsForWebsiteReferrals()
		case "addProductAndVariantToBackwardShipments":
			err = addProductAndVariantToBackwardShipments()
		case "fixCollagenCoupon":
			err = fixCollagenCoupon("product-241cc96b-1b27-4ab6-9594-046ce41ba88b", "1d1c34ad-1491-4160-af43-e3d97bcba66b")
		case "divideAmbassdorsIntoTestGroups":
			err = divideAmbassadorsIntoTestGroups()
		case "fulfillableOrder":
			err = markAllOldPaidOrdersAsFulfillable()
		case "imagesToMedia":
			err = imagesToMedia()
		default:
			err = fmt.Errorf("no case found")
		}

		fmt.Println("SCRIPT ERR", err)
	}

	if err != nil {
		panic(err)
	}
}
