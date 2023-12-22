package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetMentorReviewsPipeline(idToFind primitive.ObjectID) bson.A {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"_id", idToFind}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "reviews"},
					{"localField", "_id"},
					{"foreignField", "mentorId"},
					{"as", "reviews"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$reviews"}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "users"},
					{"localField", "reviews.menteeId"},
					{"foreignField", "_id"},
					{"as", "reviewerInfo"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$reviewerInfo"}}}},
		bson.D{
			{"$project",
				bson.D{
					{"id", "$user"},
					{"reviews",
						bson.D{
							{"review", "$reviews.review"},
							{"rating", "$reviews.rating"},
							{"date", "$reviews.date"},
							{"reviewer",
								bson.D{
									{"menteeId", "$reviewerInfo._id"},
									{"name", "$reviewerInfo.name"},
									{"jobTitle", "$reviewerInfo.jobTitle"},
									{"profileImage", "$reviewerInfo.profileImageId"},
								},
							},
						},
					},
				},
			},
		},
		bson.D{
			{"$group",
				bson.D{
					{"_id", "$_id"},
					{"reviews", bson.D{{"$push", "$reviews"}}},
				},
			},
		},
	}
	return pipeline
}

func GetFrontPageReviewsPipeline() mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"forFrontPage", true},
		}}},
		{{"$lookup", bson.D{
			{"from", "users"},
			{"localField", "menteeId"},
			{"foreignField", "_id"},
			{"as", "reviewerInfo"},
		}}},
		{{"$unwind", "$reviewerInfo"}},
		{{"$project", bson.D{
			{"mentorId", "$mentorId"},
			{"review", 1},
			{"rating", 1},
			{"date", 1},
			{"menteeName", "$reviewerInfo.name"},
			{"jobTitle", "$reviewerInfo.jobTitle"},
			{"menteeId", "$reviewerInfo._id"},
		}}},
	}
	return pipeline
}

func GetImageForUserPipeline(idToFind primitive.ObjectID) bson.A {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"_id", idToFind}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "fs.files"},
					{"localField", "profileImageId"},
					{"foreignField", "_id"},
					{"as", "profileImage"},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "fs.chunks"},
					{"localField", "profileImage._id"},
					{"foreignField", "files_id"},
					{"as", "profileImageData"},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"_id", 0},
					{"userId", "$_id"},
					{"name", "$name"},
					{"image", bson.D{{"$first", "$profileImageData.data"}}},
					{"extension", bson.D{{"$first", "$profileImage.metadata.extension"}}},
				},
			},
		},
	}
	return pipeline
}

func GetImagesForUsersPipeline(idsToFind []primitive.ObjectID) bson.A {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"_id", bson.D{{"$in", idsToFind}}}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "fs.files"},
					{"localField", "profileImageId"},
					{"foreignField", "_id"},
					{"as", "profileImage"},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "fs.chunks"},
					{"localField", "profileImage._id"},
					{"foreignField", "files_id"},
					{"as", "profileImageData"},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"_id", 0},
					{"userId", "$_id"},
					{"image", bson.D{{"$first", "$profileImageData.data"}}},
					{"extension", bson.D{{"$first", "$profileImage.metadata.extension"}}},
					{"name", "$name"},
				},
			},
		},
	}

	return pipeline
}
