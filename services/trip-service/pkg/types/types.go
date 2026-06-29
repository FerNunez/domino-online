package types

import pb "rebu/shared/proto/trip"

// OsrmApiResponse is the JSON response from the OSMR routing API
// Routes contains at least one route,
// NOTE: This service always uses Routes[0]
type OsrmApiResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

// ToProto converts the OSRM response to the proto Route message
func (o *OsrmApiResponse) ToProto() *pb.Route {
	r := o.Routes[0]
	coordinates := make([]*pb.Coordinate, len(r.Geometry.Coordinates))
	for i, c := range r.Geometry.Coordinates {
		coordinates[i] = &pb.Coordinate{
			Latitude:  c[0],
			Longitude: c[1],
		}
	}

	return &pb.Route{
		Geometry: []*pb.Geometry{{Coordinates: coordinates}},
		Distance: r.Distance,
		Duration: r.Duration,
	}
}

// PricingConfig holds the per-unit fare multiplier
type PricingConfig struct {
	PricePerUnitOfDistance float64 // cents per km
	PricePerMinute         float64 // cents per minute
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.5,
		PricePerMinute:         0.25,
	}
}
