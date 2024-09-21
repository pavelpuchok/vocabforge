package users

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
	col := db.Collection("users")
	return MongoRepository{
		col,
	}
}

type entity struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

// func entityFromModel(u models.User) (entity, error) {
// 	var id primitive.ObjectID
// 	if u.ID != "" {
// 		parsedId, err := primitive.ObjectIDFromHex(u.ID.String())
// 		if err != nil {
// 			return entity{}, fmt.Errorf("unable to create ObjectID from hex value %s. %w", u.ID, err)
// 		}
// 		id = parsedId
// 	}

// 	return entity{
// 		ID: id,
// 	}, nil
// }

func entityToModel(e entity) models.User {
	return models.User{
		ID: models.UserID(e.ID.Hex()),
	}
}

func (r MongoRepository) Create(ctx context.Context) (models.User, error) {
	insRes, err := r.col.InsertOne(ctx, entity{})
	if err != nil {
		return models.User{}, fmt.Errorf("users.MongoRepository.CreateOrUpdate unable to insert new user. %w", err)
	}
	var res entity
	err = r.col.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: insRes.InsertedID}}).Decode(&res)
	if err != nil {
		return models.User{}, fmt.Errorf("users.MongoRepository.CreateOrUpdate unable to fetch newly created user %s. %w", insRes.InsertedID, err)
	}
	return entityToModel(res), nil
}
