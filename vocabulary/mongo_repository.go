package vocabulary

import (
	"context"
	"fmt"
	"time"

	"github.com/pavelpuchok/vocabforge/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	col *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) MongoRepository {
	col := db.Collection("vocabulary")
	return MongoRepository{
		col,
	}
}

type entity struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          primitive.ObjectID `bson:"userId,omitempty"`
	Spelling        string
	Definition      string
	Language        string
	LearnStatus     string `bson:"learnStatus"`
	LexicalCategory string `bson:"lexicalCategory"`
	AnsweredCount   uint   `bson:"answeredCount"`
	Exercises       []entityExercise
	AddedAt         time.Time `bson:"addedAt"`
	LastAskedAt     time.Time `bson:"lastAskedAt,omitempty"`
}

type entityExercise struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Sentence string
	Answered bool
}

func entityToModel(e entity) (models.Word, error) {
	var status models.LearnStatus
	if err := status.UnmarshalText(e.LearnStatus); err != nil {
		return models.Word{}, fmt.Errorf("unable to unmarshal entity's status %s. %w", e.LearnStatus, err)
	}
	var lang models.Language
	if err := lang.UnmarshalText(e.Language); err != nil {
		return models.Word{}, fmt.Errorf("unable to unmarshal entity's language %s. %w", e.Language, err)
	}

	return models.Word{
		ID:              models.WordID(e.ID.Hex()),
		UserID:          models.UserID(e.UserID.Hex()),
		Spelling:        e.Spelling,
		Definition:      e.Definition,
		Language:        lang,
		LearnStatus:     status,
		LexicalCategory: e.LexicalCategory,
		AnsweredCount:   e.AnsweredCount,
		AddedAt:         e.AddedAt,
		LastAskedAt:     e.LastAskedAt,
	}, nil
}

func (r MongoRepository) AddWord(ctx context.Context, userID models.UserID, spell, definition, lexicalCategory string, lang models.Language, sentences []string) (models.Word, error) {
	userId, err := primitive.ObjectIDFromHex(userID.String())
	if err != nil {
		return models.Word{}, fmt.Errorf("vocabulary.MongoRepository.AddWord unable to build ObjectId from user's ID %s. %w", userID, err)
	}

	defStatus := models.Pending
	statusMarshalled, _ := defStatus.MarshalText()

	langMarshalled, err := lang.MarshalText()
	if err != nil {
		return models.Word{}, fmt.Errorf("vocabulary.MongoRepository.AddWord unable to marhal language %v. %w", lang, err)
	}

	exercises := make([]entityExercise, len(sentences))
	for i, text := range sentences {
		exercises[i] = entityExercise{
			ID:       primitive.NewObjectID(),
			Sentence: text,
		}
	}

	newEntity := entity{
		UserID:          userId,
		Spelling:        spell,
		Definition:      definition,
		Language:        langMarshalled,
		LearnStatus:     statusMarshalled,
		LexicalCategory: lexicalCategory,
		AnsweredCount:   0,
		Exercises:       exercises,
	}

	insRes, err := r.col.InsertOne(ctx, newEntity)
	if err != nil {
		return models.Word{}, err
	}

	var insertedEntity entity
	err = r.col.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: insRes.InsertedID}}).Decode(&insertedEntity)
	if err != nil {
		return models.Word{}, fmt.Errorf("vocabulary.MongoRepository.AddWord unable to fetch inserted document. %w", err)
	}

	m, err := entityToModel(insertedEntity)
	if err != nil {
		return models.Word{}, fmt.Errorf("vocabulary.MongoRepository.AddWord unable to map entity to model. %w", err)
	}

	return m, nil
}

func (r MongoRepository) StatsByUser(ctx context.Context, userID models.UserID) (models.SentenceExerciseStats, error) {
	userIDObj, err := primitive.ObjectIDFromHex(userID.String())
	if err != nil {
		return models.SentenceExerciseStats{}, fmt.Errorf("vocabulary.MongoRepository.StatsByUser unable to build ObjectId from user's ID %s. %w", userID, err)
	}

	p := mongo.Pipeline{
		bson.D{
			{
				Key: "$match",
				Value: bson.M{
					"userId": userIDObj,
				},
			},
		},
		bson.D{
			{
				Key: "$group",
				Value: bson.M{
					"_id": "$learnStatus",
					"count": bson.M{
						"$sum": 1,
					},
				},
			},
		},
	}

	c, err := r.col.Aggregate(ctx, p)
	if err != nil {
		return models.SentenceExerciseStats{}, fmt.Errorf("vocabulary.MongoRepository.StatsByUser unable to query stats. %w", err)
	}

	var res models.SentenceExerciseStats
	m := make(map[string]uint)
	c.Decode(&m)
	for k, v := range m {
		var status models.LearnStatus
		if err := status.UnmarshalText(k); err != nil {
			return models.SentenceExerciseStats{}, fmt.Errorf("vocabulary.MongoRepository.StatsByUser unable to parse learning status. %w", err)
		}

		switch status {
		case models.Pending:
			res.Pending = v
		case models.InProgress:
			res.InProgress = v
		case models.Learned:
			res.Learned = v
		}
	}

	return res, nil
}

func (r MongoRepository) OldestByUser(ctx context.Context, userID models.UserID, status models.LearnStatus) (models.SentenceExercise, error) {
	userIDObj, err := primitive.ObjectIDFromHex(userID.String())
	if err != nil {
		return models.SentenceExercise{}, fmt.Errorf("vocabulary.MongoRepository.OldestByUser unable to build ObjectId from user's ID %s. %w", userID, err)
	}

	opts := options.Find().SetSort(bson.D{{"lastAskedAt", 1}, {"addedAt", 1}})
	c, err := r.col.Find(ctx, bson.M{"userId": userIDObj}, opts)
	if err != nil {
		return models.SentenceExercise{}, fmt.Errorf("vocabulary.MongoRepository.OldestByUser find failed. %w", err)
	}
	defer c.Close(ctx)

	for c.Next(ctx) {

	}

	err = c.Err()
	if err != nil {
		return models.SentenceExercise{}, fmt.Errorf("vocabulary.MongoRepository.OldestByUser find cursor failed. %w", err)
	}
}
