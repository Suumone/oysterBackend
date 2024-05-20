package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"oysterProject/model"
	"time"
)

func GetMentorReviewsPipeline(idToFind primitive.ObjectID) bson.A {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"_id", idToFind}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", ReviewCollectionName},
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
					{"from", UserCollectionName},
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
			{"from", UserCollectionName},
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

func GetUserBestMentorsPipeline(idToFind primitive.ObjectID) bson.A {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"_id", idToFind}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", UserCollectionName},
					{"localField", "bestMentors"},
					{"foreignField", "_id"},
					{"as", "bestMentorsData"},
				},
			},
		},
		bson.D{{"$unwind", "$bestMentorsData"}},
		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$bestMentorsData"}}}},
	}
	return pipeline
}

func GetSessionsForNotificationPipeline(filterTimeGt, filterTimeLte time.Time) bson.A {
	pipeline := bson.A{
		bson.D{
			{"$match",
				bson.D{
					{"sessionStatus", model.Confirmed},
					{"sessionTimeStart", bson.D{
						{"$gt", filterTimeGt},
						{"$lte", filterTimeLte},
					}},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "users"},
					{"localField", "menteeId"},
					{"foreignField", "_id"},
					{"as", "mentee"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$mentee"}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "users"},
					{"localField", "mentorId"},
					{"foreignField", "_id"},
					{"as", "mentor"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$mentor"}}}},
		bson.D{
			{"$project",
				bson.D{
					{"mentorId", 1},
					{"menteeId", 1},
					{"sessionTimeStart", 1},
					{"sessionTimeEnd", 1},
					{"meetingLink", 1},
					{"paymentDetails", 1},
					{"menteeName", "$mentee.name"},
					{"menteeEmail", "$mentee.email"},
					{"mentorName", "$mentor.name"},
					{"mentorEmail", "$mentor.email"},
				},
			},
		},
	}
	return pipeline
}

func GetSessionsForReviewNotificationPipeline() bson.A {
	pipeline := bson.A{
		bson.D{
			{"$match",
				bson.D{
					{"emailWasSent", false},
					{"sessionStatus", model.Completed},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "users"},
					{"localField", "menteeId"},
					{"foreignField", "_id"},
					{"as", "mentee"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$mentee"}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "users"},
					{"localField", "mentorId"},
					{"foreignField", "_id"},
					{"as", "mentor"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$mentor"}}}},
		bson.D{
			{"$project",
				bson.D{
					{"mentorId", 1},
					{"menteeId", 1},
					{"sessionTimeStart", 1},
					{"sessionTimeEnd", 1},
					{"meetingLink", 1},
					{"paymentDetails", 1},
					{"menteeName", "$mentee.name"},
					{"menteeEmail", "$mentee.email"},
					{"mentorName", "$mentor.name"},
					{"mentorEmail", "$mentor.email"},
				},
			},
		},
	}
	return pipeline
}
