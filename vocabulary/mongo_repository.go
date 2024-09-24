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
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          primitive.ObjectID `bson:"userId,omitempty"`
	Spelling        string
	Definition      string
	Language        string
	LearnStatus     string `bson:"learnStatus"`
	LexicalCategory string `bson:"lexicalCategory"`
	AnsweredCount   uint   `bson:"answeredCount"`
	Exercises       []entityExercise
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
