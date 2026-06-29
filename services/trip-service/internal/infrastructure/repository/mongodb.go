package repository

import (
	"context"
	"fmt"
	"rebu/services/trip-service/internal/domain"
	"rebu/shared/db"
	pbd "rebu/shared/proto/driver"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	db *mongo.Database
}

func NewMongoRepository(db *mongo.Database) *mongoRepository {
	return &mongoRepository{
		db: db,
	}
}

func (r *mongoRepository) CreateTrip(ctx context.Context, t *domain.TripModel) (*domain.TripModel, error) {
	result, err := r.db.Collection(db.TripsCollection).InsertOne(ctx, t)
	if err != nil {
		return nil, err
	}
	t.ID = result.InsertedID.(primitive.ObjectID)
	return t, nil
}
func (r *mongoRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	result, err := r.db.Collection(db.RideFaresCollection).InsertOne(ctx, f)
	if err != nil {
		return err
	}
	f.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
func (r *mongoRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var fare domain.RideFareModel
	if err := r.db.Collection(db.RideFaresCollection).FindOne(ctx, bson.M{"_id": oid}).Decode(&fare); err != nil {
		return nil, err
	}
	return &fare, nil
}
func (r *mongoRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var trip domain.TripModel
	if err := r.db.Collection(db.TripsCollection).FindOne(ctx, bson.M{"_id": oid}).Decode(&trip); err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *mongoRepository) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error {
	oid, err := primitive.ObjectIDFromHex(tripID)
	if err != nil {
		return err
	}
	set := bson.M{"status": status}
	if driver != nil {
		set["driver"] = driver
	}
	result, err := r.db.Collection(db.TripsCollection).UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": set})
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("trip not found: %s", tripID)
	}
	return nil
}
