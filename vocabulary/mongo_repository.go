package vocabulary

import (
	"context"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	UserID        primitive.ObjectID `bson:"userId,omitempty"`
	Spelling      string
	Definition    string
	Language      string
	LearnStatus   string
	AnsweredCount uint
	Exercises     []models.SentenceExercise
}

func entityToModel(e entity) (models.Word, error) {
	var status models.LearnStatus
	if err := status.UnmarshalText(e.LearnStatus); err != nil {
		return models.Word{}, fmt.Errorf("unable to unmarshal entity's status %s. %w", e.LearnStatus, err)
	}

	return models.Word{
		ID:            models.WordID(e.ID.Hex()),
		UserID:        models.UserID(e.UserID.Hex()),
		Spelling:      e.Spelling,
		Definition:    e.Definition,
		Language:      models.Language(e.Language),
		LearnStatus:   status,
		AnsweredCount: e.AnsweredCount,
	}, nil
}

func (r MongoRepository) AddWord(ctx context.Context, userID models.UserID, spell, definition, lang string, exercises []models.SentenceExercise) (models.Word, error) {
	userId, err := primitive.ObjectIDFromHex(userID.String())
	if err != nil {
		return models.Word{}, fmt.Errorf("vocabulary.MongoRepository.AddWord unable to build ObjectId from user's ID %s. %w", userID, err)
	}

	status := models.Pending

	newEntity := entity{
		UserID:        userId,
		Spelling:      spell,
		Definition:    definition,
		Language:      lang,
		LearnStatus:   status.String(),
		AnsweredCount: 0,
		Exercises:     exercises,
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
