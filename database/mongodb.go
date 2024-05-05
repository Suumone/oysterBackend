package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/url"
	"oysterProject/model"
	"oysterProject/utils"
	"strconv"
	"strings"
)

func CreateMentor(user *model.User) (primitive.ObjectID, error) {
	collection := GetCollection(UserCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	doc, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error inserting user: %v\n", err)
		return primitive.ObjectID{}, err
	}
	log.Printf("User(name: %s, insertedID: %s) inserted successfully\n", user.Username, doc.InsertedID)
	return doc.InsertedID.(primitive.ObjectID), nil
}

func GetMentors(params url.Values, userId primitive.ObjectID) ([]*model.User, error) {
	filter, err := getFilterForMentorList(params, userId)
	if err != nil {
		return nil, err
	}
	offset, limit, err := getOffsetAndLimit(params)
	if err != nil {
		return nil, err
	}
	return fetchMentors(filter, offset, limit, nil)
}

func GetTopMentors(params url.Values) ([]*model.User, error) {
	filter := getFilterForTopMentorList()
	offset, limit, err := getOffsetAndLimit(params)
	if err != nil {
		return nil, err
	}
	sortBson := bson.D{{"topMentorOrder", 1}}
	return fetchMentors(filter, offset, limit, sortBson)
}

func getOffsetAndLimit(params url.Values) (int, int, error) {
	var offset int
	var err error
	if params.Get("offset") != "" {
		offset, err = strconv.Atoi(params.Get("offset"))
		if err != nil {
			log.Printf("Error reading offset parameter: %v\n\n", err)
			return 0, 0, err
		}
	}
	var limit int
	if params.Get("limit") != "" {
		limit, err = strconv.Atoi(params.Get("limit"))
		if err != nil {
			log.Printf("Error reading limit parameter: %v\n\n", err)
			return 0, 0, err
		}
	}
	return offset, limit, nil
}

func getFilterForMentorList(params url.Values, userId primitive.ObjectID) (bson.M, error) {
	filter := bson.M{
		"isApproved": true,
	}

	for key, values := range params {
		fieldType, _ := fieldTypes[key]
		switch fieldType {
		case "options":
			continue
		case "array":
			filter[key] = bson.M{"$in": strings.Split(values[0], ",")}
		case "number":
			filter[key] = bson.M{"$gt": convertStringToNumber(values[0])}
		default:
			filter[key] = bson.M{"$regex": values[0], "$options": "i"}
		}
	}
	if !userId.IsZero() {
		bestMentors, err := getUserBestMentorIds(userId)
		if err != nil {
			return nil, err
		}
		if len(bestMentors) > 0 {
			filter["_id"] = bson.M{"$nin": bestMentors}
		}
	}
	log.Printf("MongoDB filter:%s\n", filter)
	return filter, nil
}

func getFilterForTopMentorList() bson.M {
	return bson.M{
		"isApproved":  true,
		"isTopMentor": true,
	}
}

func fetchMentors(filter bson.M, offset int, limit int, sortBson bson.D) ([]*model.User, error) {
	collection := GetCollection(UserCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	opts := options.Find()
	if offset != 0 {
		opts = opts.SetSkip(int64(offset))
	}
	if limit != 0 {
		opts = opts.SetLimit(int64(limit))
	}
	if sortBson != nil {
		opts.SetSort(sortBson)
	}
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Failed to find documents: %v\n", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode document: %v", err)
			return nil, err
		} else {
			users = append(users, &user)
		}
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		return nil, err
	}
	return users, nil
}

func GetUserWithImageByID(id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()

	userChan := make(chan *model.User)
	errChan := make(chan error)
	go func() {
		userImage, err := GetUserPictureByUserId(id)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- &model.User{UserImage: userImage}
	}()

	collection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": id}
	var user model.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		handleFindError(err, id.Hex(), "user")
		return nil, err
	}

	select {
	case userFromChan := <-userChan:
		user.UserImage = userFromChan.UserImage
	case errFromChan := <-errChan:
		if errors.Is(errFromChan, utils.UserImageNotFound) {
			log.Printf("GetUserWithImageByID: Failed to find image for user(%s) error: %v\n", id, errFromChan)
		} else if errFromChan != nil {
			log.Printf("GetUserWithImageByID: Search image for user(%s) error: %v\n", id, errFromChan)
			return nil, errFromChan
		}
	}
	return &user, err
}

func GetUserByID(id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": id}
	var user model.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		handleFindError(err, id.Hex(), "user")
		return nil, err
	}
	return &user, err
}

func GetMentorReviewsByID(id string) (*model.UserWithReviews, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection(UserCollectionName)
	idToFind, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("GetMentorReviewsByID: Failed to convert mentor id(%s): %v", id, err)
		return nil, err
	}

	mentorListPipeline := GetMentorReviewsPipeline(idToFind)
	cursor, err := usersColl.Aggregate(ctx, mentorListPipeline)
	if err != nil {
		log.Printf("GetMentorReviewsByID: Failed to decode document: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var user model.UserWithReviews
	for cursor.Next(ctx) {
		if err := cursor.Decode(&user); err != nil {
			log.Printf("GetMentorReviewsByID: Failed to decode document: %v", err)
			return nil, err
		}
	}

	reviewerIDs := extractReviewerIDs(user.Reviews)
	userImagesMap := make(map[primitive.ObjectID]*model.UserImage)
	if len(reviewerIDs) > 0 {
		usersWithImages, err := GetUserImages(reviewerIDs)
		if err != nil {
			log.Printf("GetMentorReviewsByID: Failed to get user images: %v", err)
			return nil, err
		}
		for _, userImage := range usersWithImages {
			userImagesMap[userImage.UserId] = userImage
		}
	}

	updateUserReviews(user, userImagesMap)
	return &user, nil
}

func extractReviewerIDs(reviews []model.Reviews) []primitive.ObjectID {
	var reviewerIDs []primitive.ObjectID
	for _, review := range reviews {
		reviewerIDs = append(reviewerIDs, review.Reviewer.MenteeId)
	}
	return reviewerIDs
}

func updateUserReviews(user model.UserWithReviews, userImagesMap map[primitive.ObjectID]*model.UserImage) {
	for i, review := range user.Reviews {
		if userImage, ok := userImagesMap[review.Reviewer.MenteeId]; ok {
			user.Reviews[i].Reviewer.UserImage = userImage
		}
	}
}

func UpdateAndGetUser(user *model.User, id primitive.ObjectID) (*model.User, error) {
	userChan := make(chan *model.User)
	errChan := make(chan error)
	go func() {
		userWithImage, err := GetUserWithImageByID(id)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- userWithImage
	}()

	collection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": id}
	updateOp := bson.M{"$set": user}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err := collection.UpdateOne(ctx, filter, updateOp)
	if err != nil {
		return nil, err
	}

	var userAfterUpdate *model.User
	select {
	case userFromChan := <-userChan:
		userAfterUpdate = userFromChan
	case errFromChan := <-errChan:
		return nil, errFromChan
	}
	log.Printf("User(id: %s) updated successfully!\n", id)
	return userAfterUpdate, nil
}

func GetListOfFilterFields() ([]map[string]interface{}, error) {
	var fields []map[string]interface{}
	filterColl := GetCollection(FieldInfoCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	cursor, err := filterColl.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("Error executing filter fields search in db: %v", err)
		return nil, err
	}

	var metas []map[string]interface{}
	if err = cursor.All(context.TODO(), &metas); err != nil {
		return nil, err
	}

	for _, meta := range metas {
		fieldData, err := extractFieldDataFromMeta(meta)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldData)
	}

	return fields, nil
}

func GetFiltersByNames(params model.RequestParams) ([]map[string]interface{}, error) {
	var fields []map[string]interface{}
	filter := bson.M{
		"fieldStorage": bson.M{"$in": params.Fields},
	}
	filterColl := GetCollection(FieldInfoCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	cursor, err := filterColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var metas []map[string]interface{}
	if err = cursor.All(context.TODO(), &metas); err != nil {
		return nil, err
	}

	for _, meta := range metas {
		fieldData, err := extractFieldDataFromMeta(meta)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldData)
	}

	return fields, nil
}

func extractFieldDataFromMeta(meta map[string]interface{}) (map[string]interface{}, error) {
	fieldName := meta["fieldName"].(string)
	fieldType := meta["type"].(string)
	fieldStorage := meta["fieldStorage"].(string)

	valuesFromDb, ok := meta["values"].(primitive.A)
	if !ok {
		valuesFromDb = primitive.A{}
	}

	fieldData := map[string]interface{}{
		"fieldName":    fieldName,
		"type":         fieldType,
		"fieldStorage": fieldStorage,
		"values":       valuesFromDb,
	}

	if fieldType == "dropdown" && len(valuesFromDb) == 0 {
		usersColl := GetCollection(UserCollectionName)
		values, err := usersColl.Distinct(context.TODO(), fieldStorage, bson.D{})
		if err != nil {
			log.Printf("Error executing distinct search in db: %v", err)
			return nil, err
		}
		fieldData["values"] = values
	}

	return fieldData, nil
}

func GetReviewsForFrontPage() ([]*model.ReviewsForFrontPage, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	reviewColl := GetCollection(ReviewCollectionName)
	pipeline := GetFrontPageReviewsPipeline()
	cursor, err := reviewColl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("GetReviewsForFrontPage: Failed to create reviews cursor: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*model.ReviewsForFrontPage
	for cursor.Next(ctx) {
		var review model.ReviewsForFrontPage
		if err := cursor.Decode(&review); err != nil {
			log.Printf("GetReviewsForFrontPage: Failed to decode reviews: %v", err)
			return nil, err
		}
		result = append(result, &review)
	}
	var reviewerIDs []primitive.ObjectID
	for _, review := range result {
		reviewerIDs = append(reviewerIDs, review.MenteeId)
	}
	userImagesMap := make(map[primitive.ObjectID]*model.UserImage)

	if len(reviewerIDs) > 0 {
		usersWithImages, err := GetUserImages(reviewerIDs)
		if err != nil {
			log.Printf("GetReviewsForFrontPage: Failed to find image for users(%s): %v\n", reviewerIDs, err)
			return nil, err
		}
		for _, userImage := range usersWithImages {
			userImagesMap[userImage.UserId] = userImage
		}
	}

	for i, review := range result {
		if userImage, ok := userImagesMap[review.MenteeId]; ok {
			result[i].MenteeImage = userImage
		}
	}

	return result, nil
}

func GetUserByEmail(email string) (*model.User, error) {
	usersCollection := GetCollection(UserCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	filter := bson.M{"email": email}
	var user model.User
	err := usersCollection.FindOne(ctx, filter).Decode(&user)
	return &user, err
}

func ChangePassword(userId primitive.ObjectID, passwordPayload model.PasswordChange) error {
	userCollection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}
	var user model.User
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	err := userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Printf("Failed to find document: %v\n", err)
		return err
	}
	if checkPassword(user.Password, passwordPayload.OldPassword) {
		return updatePassword(userId, passwordPayload.NewPassword)
	} else {
		return errors.New("old passwords do not match")
	}
}

func checkPassword(hashedPassword string, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return false
	}
	return true
}

func updatePassword(userId primitive.ObjectID, plainPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	userCollection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}
	update := bson.M{
		"$set": bson.M{
			"password": hashedPassword,
		},
	}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func GetCurrentState(userId primitive.ObjectID) (*model.UserState, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}
	var user model.UserState
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		handleFindError(err, userId.Hex(), "user")
		return nil, err
	}
	return &user, nil
}

func UpdateUserState(userId primitive.ObjectID) error {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}
	var user model.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		handleFindError(err, userId.Hex(), "user")
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"asMentor": !user.AsMentor,
		},
	}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update user(%s) state: %v\n", userId, err)
		return err
	}

	log.Printf("User(id: %s) state updated successfully!\n", userId)
	return nil
}

func SaveProfilePicture(userId primitive.ObjectID, fileBytes []byte, fileExtension string) error {
	bucket, err := gridfs.NewBucket(
		MongoDBOyster,
	)
	if err != nil {
		log.Printf("Failed create bucket for user (id: %s) error:%s\n", userId, err)
		return err
	}
	uploadStream, err := bucket.OpenUploadStream(userId.Hex()+"_picture", options.GridFSUpload().SetMetadata(bson.M{"extension": fileExtension}).SetChunkSizeBytes(utils.ImageLimitSizeMB))
	if err != nil {
		log.Printf("Failed to open image stream for user (id: %s) error:%s\n", userId, err)
		return err
	}
	defer uploadStream.Close()
	_, err = uploadStream.Write(fileBytes)
	if err != nil {
		log.Printf("Failed to upload image for user (id: %s) error:%s\n", userId, err)
		return err
	}

	userCollection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}
	update := bson.M{
		"$set": bson.M{
			"profileImageId": uploadStream.FileID.(primitive.ObjectID),
		},
	}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to upadet profileImageId for user (id: %s) error:%s\n", userId, err)
		return err
	}
	return nil
}

func GetUserPictureByUserId(userId primitive.ObjectID) (*model.UserImage, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection(UserCollectionName)
	imageForUserPipeline := GetImageForUserPipeline(userId)
	cursor, err := usersColl.Aggregate(ctx, imageForUserPipeline)
	if err != nil {
		log.Printf("Failed to execute image search: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var userImage model.UserImage
	for cursor.Next(ctx) {
		if err := cursor.Decode(&userImage); err != nil {
			log.Printf("Failed to decode image search result: %v", err)
			return nil, err
		}
	}
	if utils.IsEmptyStruct(userImage) || len(userImage.Image) == 0 {
		return nil, utils.UserImageNotFound
	}
	return &userImage, nil
}

func SaveBestMentorsForUser(userId primitive.ObjectID, mentors []model.MentorForRequest) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}

	var mentorsIds []string
	for _, mentor := range mentors {
		mentorsIds = append(mentorsIds, mentor.MentorId)
	}
	mentorsObjectIds, err := convertStringsToObjectIDs(mentorsIds)
	if err != nil {
		return
	}
	updateOp := bson.M{"$set": bson.M{"bestMentors": mentorsObjectIds}}
	_, err = usersColl.UpdateOne(ctx, filter, updateOp)
	if err != nil {
		log.Printf("Failed to save mentors from chatgpt to db: %v", err)
	} else {
		log.Printf("Mentors(%s) for user(%s) saved in db", mentorsIds, userId)
	}
}

func getUserBestMentorIds(userId primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection(UserCollectionName)
	filter := bson.M{"_id": userId}
	var user model.UserBestMentors
	err := usersColl.FindOne(ctx, filter).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Printf("User(%s) was not found in db: %v\n", userId, err)
		return nil, err
	} else if err != nil {
		log.Printf("Failed to get user(%s) best mentors from db: %v\n", userId, err)
		return nil, err
	}
	return user.BestMentors, nil
}

func GetBestMentors(userId primitive.ObjectID) ([]*model.User, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(UserCollectionName)
	pipeline := GetUserBestMentorsPipeline(userId)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("GetBestMentors: failed to aggregate best mentors: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*model.User
	for cursor.Next(ctx) {
		var user model.User
		err = cursor.Decode(&user)
		if err != nil {
			log.Printf("GetBestMentors: failed to decode user: %v", err)
			return nil, err
		}
		result = append(result, &user)
	}
	var userIds []primitive.ObjectID
	for _, user := range result {
		userIds = append(userIds, user.Id)
	}
	usersWithImages, err := GetUserImages(userIds)
	if err != nil {
		log.Printf("GetBestMentors: Failed to get user images: %v", err)
		return nil, err
	}
	for _, userImage := range usersWithImages {
		for i, user := range result {
			if userImage.UserId == user.Id {
				result[i].UserImage = userImage
			}
		}
	}

	return result, nil
}

func GetValuesForSelect(params url.Values) ([]model.ValuesToSelect, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(ValuesForSelectCollectionName)
	filterArray := strings.Split(params.Get("fields"), ",")
	var cursor *mongo.Cursor
	var err error
	if len(filterArray) > 0 && filterArray[0] != "" {
		filter := bson.M{
			"name": bson.M{"$in": filterArray},
		}
		cursor, err = collection.Find(ctx, filter)
	} else {
		cursor, err = collection.Find(ctx, bson.D{})
	}

	if err != nil {
		log.Printf("Failed to get values from collection selectValues from DB: %v\n", err)
		return nil, err
	}

	var valuesToSelect []model.ValuesToSelect
	if err = cursor.All(context.TODO(), &valuesToSelect); err != nil {
		log.Printf("Failed to map values from collection selectValues from DB: %v\n", err)
		return nil, err
	}
	return valuesToSelect, nil
}

func UpdateMentorRequest(request string, id primitive.ObjectID) {
	collection := GetCollection(UserCollectionName)
	filter := bson.M{"_id": id}
	updateOp := bson.M{"$set": bson.M{"userMentorRequest": request}}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err := collection.UpdateOne(ctx, filter, updateOp)
	if err != nil {
		log.Printf("Failed to update mentor request(%s) for user(%s) in DB: %v\n", request, id.Hex(), err)
		return
	}

	log.Printf("Mentor request for user(id: %s) updated successfully!\n", id.Hex())
}

func GetUserImages(userIds []primitive.ObjectID) ([]*model.UserImage, error) {
	if len(userIds) == 0 {
		log.Println("GetUserImages: empty list of users")
		return nil, nil
	}

	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	imagesForUsersPipeline := GetImagesForUsersPipeline(userIds)
	usersColl := GetCollection(UserCollectionName)
	cursor, err := usersColl.Aggregate(ctx, imagesForUsersPipeline)
	if err != nil {
		log.Printf("Failed to aggregate user with image: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*model.UserImage
	for cursor.Next(ctx) {
		var user model.UserImage
		err := cursor.Decode(&user)
		if err != nil {
			log.Printf("Failed to decode user with image: %v", err)
			return nil, err
		}
		result = append(result, &user)
	}
	return result, nil
}
