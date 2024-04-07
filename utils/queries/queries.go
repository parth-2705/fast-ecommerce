package utils

import (
	"go.mongodb.org/mongo-driver/bson"
)

var ProductQuery bson.A = bson.A{bson.D{
	{Key: "$lookup",
		Value: bson.D{
			{Key: "from", Value: "brands"},
			{Key: "localField", Value: "brandID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "brand"},
		},
	}},
	bson.D{
		{Key: "$unwind",
			Value: bson.D{
				{Key: "path", Value: "$brand"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			},
		},
	},
	// bson.D{
	// 	{Key: "$lookup",
	// 		Value: bson.D{
	// 			{Key: "from", Value: "variants"},
	// 			{Key: "localField", Value: "_id"},
	// 			{Key: "foreignField", Value: "productID"},
	// 			{Key: "as", Value: "variants"},
	// 		},
	// 	},
	// }
}

var PageViewQuery bson.A = bson.A{bson.D{
	{Key: "$lookup",
		Value: bson.D{
			{Key: "from", Value: "products"},
			{Key: "localField", Value: "resourceID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "resource"},
		},
	}},
	bson.D{
		{Key: "$unwind",
			Value: bson.D{
				{Key: "path", Value: "$resource"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			},
		},
	},
}

var CategoriesPageQuery bson.A = bson.A{
	bson.D{
		{"$lookup",
			bson.D{
				{"from", "categories"},
				{"localField", "_id"},
				{"foreignField", "parentID"},
				{"as", "childrenCategories"},
			},
		},
	},
}

var CategoriesChildrenIDQuery bson.A = bson.A{
	bson.D{
		{"$graphLookup",
			bson.D{
				{"from", "categories"},
				{"startWith", "$_id"},
				{"connectFromField", "_id"},
				{"connectToField", "parentID"},
				{"as", "childrenCategories"},
			},
		},
	},
	bson.D{
		{"$group",
			bson.D{
				{"_id", nil},
				{"childrenID", bson.D{{"$addToSet", "$childrenCategories._id"}}},
			},
		},
	},
	bson.D{
		{"$unwind",
			bson.D{
				{"path", "$childrenID"},
				{"preserveNullAndEmptyArrays", false},
			},
		},
	},
}

var FullCategoryQuery bson.A = bson.A{
	bson.D{
		{"$lookup",
			bson.D{
				{"from", "categories"},
				{"localField", "parentID"},
				{"foreignField", "_id"},
				{"as", "parent"},
			},
		},
	},
	bson.D{
		{"$unwind",
			bson.D{
				{"path", "$parent"},
				{"preserveNullAndEmptyArrays", true},
			},
		},
	},
}

func GetBrandCategoriesByID(brandID string) (pipeline bson.A) {
	pipeline = bson.A{
		bson.D{{"$match", bson.D{{"_id", brandID}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "products"},
					{"localField", "_id"},
					{"foreignField", "brandID"},
					{"as", "products"},
				},
			},
		},
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$products"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		bson.D{
			{"$group",
				bson.D{
					{"_id", "test"},
					{"category", bson.D{{"$addToSet", "$products.newCategory"}}},
				},
			},
		},
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$category"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "categories"},
					{"localField", "category"},
					{"foreignField", "_id"},
					{"as", "category"},
				},
			},
		},
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$category"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
	}
	return
}

func QueryGetAllAppliedCampaignsByInfluencer(influencerID string) (pipeline bson.A) {
	pipeline = bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "influencerID", Value: influencerID}}}},
	}

	pipeline = append(pipeline, QueryGetAllAppliedCampaigns()...)
	return pipeline
}

func QueryGetAllAppliedCampaigns() (pipeline bson.A) {
	return bson.A{
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "campaigns"},
					{Key: "localField", Value: "campaignID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "campaign"},
				},
			},
		},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$campaign"}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "brands"},
					{Key: "localField", Value: "campaign.brandID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "campaign.brand"},
				},
			},
		},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$campaign.brand"}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "products"},
					{Key: "localField", Value: "campaign.products"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "campaign.productArray"},
				},
			},
		},
		// bson.D{
		// 	{Key: "$project",
		// 		Value: bson.D{
		// 			{Key: "_id", Value: 0},
		// 			{Key: "campaign", Value: 1},
		// 		},
		// 	},
		// },
	}
}

func GetAllCampaigns() (pipeline bson.A) {
	return bson.A{
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "brands"},
					{Key: "localField", Value: "brandID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "brand"},
				},
			}},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$brand"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "products"},
					{Key: "localField", Value: "products"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "productArray"},
				},
			}},
	}
}
