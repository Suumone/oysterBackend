package database

import (
	"context"
	"encoding/base64"
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

func CreateMentor(user model.User) (primitive.ObjectID, error) {
	collection := GetCollection("users")
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

func GetMentors(params url.Values, userId string) ([]model.User, error) {
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

func GetTopMentors(params url.Values) ([]model.User, error) {
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

func getFilterForMentorList(params url.Values, userId string) (bson.M, error) {
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
	if userId != "" {
		bestMentors, err := getUserBestMentors(userId)
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

func fetchMentors(filter bson.M, offset int, limit int, sortBson bson.D) ([]model.User, error) {
	collection := GetCollection("users")
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
	defer cursor.Close(context.Background())

	var users []model.User
	for cursor.Next(context.Background()) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode document: %v", err)
			return nil, err
		} else {
			users = append(users, user)
		}
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		return nil, err
	}
	return users, nil
}

func GetUserByID(id string) model.User {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": idToFind}
	var user model.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		handleFindError(err, id)
		return model.User{}
	}

	userImage, err := GetUserPictureByUserId(id)
	if err != nil && !errors.Is(err, utils.UserImageNotFound) {
		log.Printf("Failed to find image for user(%s): %v\n", id, err)
	}
	user.UserImage = userImage
	return user
}

func GetMentorReviewsByID(id string) model.UserWithReviews {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)

	mentorListPipeline := GetMentorListPipeline(idToFind)
	cursor, err := usersColl.Aggregate(ctx, mentorListPipeline)
	if err != nil {
		return model.UserWithReviews{}
	}
	defer cursor.Close(ctx)
	var user model.UserWithReviews
	for cursor.Next(context.Background()) {
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode document: %v", err)
		}
	}

	return user
}

func UpdateUser(user model.User, id string) (model.User, error) {
	user.IsNewUser = false
	idToFind, _ := primitive.ObjectIDFromHex(id)
	collection := GetCollection("users")
	filter := bson.M{"_id": idToFind}
	updateOp := bson.M{"$set": user}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err := collection.UpdateOne(ctx, filter, updateOp)
	if err != nil {
		return model.User{}, err
	}
	userAfterUpdate := GetUserByID(id)
	log.Printf("User(id: %s) updated successfully!\n", id)
	return userAfterUpdate, nil
}

func GetListOfFilterFields() ([]map[string]interface{}, error) {
	var fields []map[string]interface{}
	filterColl := GetCollection("fieldInfo")
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	cursor, err := filterColl.Find(ctx, bson.D{})
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

func GetFiltersByNames(params model.RequestParams) ([]map[string]interface{}, error) {
	var fields []map[string]interface{}
	filter := bson.M{
		"fieldStorage": bson.M{"$in": params.Fields},
	}
	filterColl := GetCollection("fieldInfo")
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
		usersColl := GetCollection("users")
		values, err := usersColl.Distinct(context.TODO(), fieldStorage, bson.D{})
		if err != nil {
			return nil, err
		}
		fieldData["values"] = values
	}

	return fieldData, nil
}

func GetReviewsForFrontPage() []model.ReviewsForFrontPage {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	reviewColl := GetCollection("reviews")
	pipeline := GetFrontPageReviewsPipeline()
	cursor, err := reviewColl.Aggregate(ctx, pipeline)
	if err != nil {
		return []model.ReviewsForFrontPage{}
	}
	defer cursor.Close(ctx)

	var result []model.ReviewsForFrontPage
	for cursor.Next(ctx) {
		var review model.ReviewsForFrontPage
		err := cursor.Decode(&review)
		if err != nil {
			return []model.ReviewsForFrontPage{}
		}
		result = append(result, review)
	}
	return result
}

func GetUserByEmail(email string) (model.User, error) {
	usersCollection := GetCollection("users")
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	filter := bson.M{"email": email}
	var user model.User
	err := usersCollection.FindOne(ctx, filter).Decode(&user)
	return user, err
}

func ChangePassword(userId string, passwordPayload model.PasswordChange) error {
	userCollection := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
	var user model.User
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	err := userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Printf("Failed to find document: %v\n", err)
		return err
	}
	if checkPassword(user.Password, passwordPayload.OldPassword) {
		return updatePassword(idToFind, passwordPayload.NewPassword)
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
	userCollection := GetCollection("users")
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

func GetCurrentState(userId string) model.UserState {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
	var user model.UserState
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("User(%s) not found\n", userId)
		} else {
			log.Printf("Failed to find document: %v\n", err)
		}
		return model.UserState{}
	}
	return user
}

func UpdateUserState(userId string) error {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
	var user model.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		handleFindError(err, userId)
		return err
	}

	filter = bson.M{"_id": idToFind}
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

func SaveProfilePicture(userId string, fileBytes []byte, fileExtension string) error {
	bucket, err := gridfs.NewBucket(
		MongoDBOyster,
	)
	if err != nil {
		return err
	}
	uploadStream, err := bucket.OpenUploadStream(userId+"_picture", options.GridFSUpload().SetMetadata(bson.M{"extension": fileExtension}).SetChunkSizeBytes(utils.ImageLimitSizeMB))
	if err != nil {
		return err
	}
	defer uploadStream.Close()
	_, err = uploadStream.Write(fileBytes)
	if err != nil {
		return err
	}

	userCollection := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
	update := bson.M{
		"$set": bson.M{
			"profileImageId": uploadStream.FileID.(primitive.ObjectID),
		},
	}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err = userCollection.UpdateOne(ctx, filter, update)
	return err
}

func GetUserPictureByUserId(userId string) (model.UserImageResult, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	imageForUserPipeline := GetImageForUserPipeline(idToFind)
	cursor, err := usersColl.Aggregate(ctx, imageForUserPipeline)
	if err != nil {
		log.Printf("Failed to execute image search: %v", err)
		return model.UserImageResult{}, err
	}
	defer cursor.Close(ctx)
	var userImage model.UserImage
	for cursor.Next(context.Background()) {
		if err := cursor.Decode(&userImage); err != nil {
			log.Printf("Failed to decode image search result: %v", err)
			return model.UserImageResult{}, err
		}
	}
	if utils.IsEmptyStruct(userImage) || len(userImage.Image) == 0 {
		return model.UserImageResult{}, utils.UserImageNotFound
	}
	userImageResult := model.UserImageResult{
		UserId:    userImage.UserId,
		Image:     base64.StdEncoding.EncodeToString(userImage.Image[0]),
		Extension: userImage.Extension,
	}
	return userImageResult, nil
}

func SaveBestMentorsForUser(userId string, mentorsIds []string) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
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

func getUserBestMentors(userId string) ([]primitive.ObjectID, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
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

func GetValuesForSelect(params url.Values) ([]model.ValuesToSelect, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("selectValues")
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

func UpdateMentorRequest(request string, id string) {
	idToFind, _ := primitive.ObjectIDFromHex(id)
	collection := GetCollection("users")
	filter := bson.M{"_id": idToFind}
	updateOp := bson.M{"$set": bson.M{"userMentorRequest": request}}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err := collection.UpdateOne(ctx, filter, updateOp)
	if err != nil {
		log.Printf("Failed to update mentor request(%s) for user(%s) in DB: %v\n", request, id, err)
		return
	}

	log.Printf("Mentor request for user(id: %s) updated successfully!\n", id)
}

func GetTotalUserExperience(userId string) (float32, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	idToFind, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": idToFind}
	opts := options.FindOne().SetProjection(bson.M{"areaOfExpertise": 1})
	collection := GetCollection("users")
	var user model.User
	err := collection.FindOne(ctx, filter, opts).Decode(&user)
	if err != nil {
		handleFindError(err, userId)
		return -1, err
	}
	var totalExperience float32 = 0
	for _, entry := range user.AreaOfExpertise {
		totalExperience += entry.Experience
	}
	return totalExperience, nil
}
