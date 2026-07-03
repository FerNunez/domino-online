package domain

import (
	"rebu/services/trip-service/pkg/types"
	pb "rebu/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RideFareModel is the document model for a fare quote
// One document is created per vehicle catefory during PreviewTrip
// It embeds the full Route so CreateTrip can attach it to the trip without a seconde db lookup
type RideFareModel struct {
	ID                primitive.ObjectID     `bson:"_id,omitempty"`
	UserID            string                 `bson:"userID"`
	PackageSlut       string                 `bson:"packageSlut"` //suv | sedan | etc
	TotalPriceInCents float64                `bson:"totalPriceInCents"`
	Route             *types.OsrmApiResponse `bson:"route"`
}

// ToProto converts this fare to the proto representation for GRPC reponses
func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlut,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

// ToRideFaresProto converts a slice of fare models to a slice of proto fares
func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	pbFares := make([]*pb.RideFare, len(fares))
	for i, f := range fares {
		pbFares[i] = f.ToProto()
	}
	return pbFares
}
