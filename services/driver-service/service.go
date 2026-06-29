package main

import (
	math "math/rand/v2"
	"sync"

	pb "rebu/shared/proto/driver"
	"rebu/shared/util"

	"github.com/mmcloughlin/geohash"
)

type DriverInMap struct {
	Driver *pb.Driver
}

type Service struct {
	drivers []*DriverInMap
	mu      sync.RWMutex
}

func NewService() *Service {
	return &Service{
		drivers: make([]*DriverInMap, 0),
	}
}

// RegisterDriver adds a new driver to the registry with a random starting location
// drawn from PredefinedRoutes, a random car plate, and a random avatar URL
func (s *Service) RegisterDriver(driverID, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// random route
	idx := math.IntN(len(PredefinedRoutes))
	route := PredefinedRoutes[idx]

	driver := &pb.Driver{
		Id:             driverID,
		Name:           "Lando Norris",
		ProfilePicture: util.GetRandomAvatar(idx),
		CarPlate:       GenerateRandomPlate(),
		Geohash:        geohash.Encode(route[0][0], route[0][1]),
		PackageSlug:    packageSlug,
		Location: &pb.Location{
			Latitude:  route[0][0],
			Longitude: route[0][1],
		},
	}
	s.drivers = append(s.drivers, &DriverInMap{driver})
	return driver, nil
}

// UnregisterDriver removes a driver from the registry by ID
// Called when the driver's WebSocket connection closes
func (s *Service) UnregisterDriver(driverID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, d := range s.drivers {
		if d.Driver.Id == driverID {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
			return
		}
	}
}

// FindAvailableDriver returns the IDs of all driver registered for the given slug (vehicle catagory)
// Returns an empty slice if none are available
func (s *Service) FindAvailableDrivers(packageSlug string) []string {
	// No RLock needed here in production because this is called from a single
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ids []string
	for _, d := range s.drivers {
		if d.Driver.PackageSlug == packageSlug {
			ids = append(ids, d.Driver.Id)
		}
	}
	return ids
}
