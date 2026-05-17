package data

import (
	"context"
	"math"

	"gitlab.calendaria.team/services/agents/internal/conf"
	stores_v1 "gitlab.calendaria.team/services/stores/api/stores/v1"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StoresClient provides access to the stores service.
type StoresClient interface {
	GetStore(ctx context.Context, storeID int64) (*stores_v1.Store, error)
	ValidateProximity(ctx context.Context, storeID int64, agentLat, agentLon float64) (bool, error)
}

type storesClient struct {
	client stores_v1.StoresClient
	log    *log.Helper
}

type stubStoresClient struct {
	log *log.Helper
}

// NewStoresClientFromConf creates a StoresClient from config.
func NewStoresClientFromConf(bc *conf.Bootstrap, logger log.Logger) (StoresClient, func(), error) {
	endpoint := ""
	if bc.GetDiscovery() != nil {
		endpoint = bc.GetDiscovery().GetStoresEndpoint()
	}
	return NewStoresClient(endpoint, logger)
}

// NewStoresClient creates a gRPC client for the stores service.
// If endpoint is empty, returns a stub that always allows proximity.
func NewStoresClient(endpoint string, logger log.Logger) (StoresClient, func(), error) {
	l := log.NewHelper(logger)

	if endpoint == "" {
		l.Info("Stores endpoint not configured, using stub client")
		return &stubStoresClient{log: l}, func() {}, nil
	}

	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := stores_v1.NewStoresClient(conn)
	l.Infof("Connected to stores service at %s", endpoint)

	cleanup := func() {
		conn.Close()
	}

	return &storesClient{client: client, log: l}, cleanup, nil
}

func (c *storesClient) GetStore(ctx context.Context, storeID int64) (*stores_v1.Store, error) {
	reply, err := c.client.GetStore(ctx, &stores_v1.StoreRequest{StoreId: storeID})
	if err != nil {
		return nil, err
	}
	return reply.GetStore(), nil
}

// ValidateProximity checks if the agent is within 200m of the store.
func (c *storesClient) ValidateProximity(ctx context.Context, storeID int64, agentLat, agentLon float64) (bool, error) {
	store, err := c.GetStore(ctx, storeID)
	if err != nil {
		return false, err
	}

	if store.Lat == nil || store.Lon == nil {
		// Store has no coordinates — allow check-in
		return true, nil
	}

	distance := haversineDistance(agentLat, agentLon, *store.Lat, *store.Lon)
	return distance <= 200.0, nil
}

// Stub implementation — always allows proximity and returns nil store.
func (c *stubStoresClient) GetStore(_ context.Context, _ int64) (*stores_v1.Store, error) {
	return &stores_v1.Store{}, nil
}

func (c *stubStoresClient) ValidateProximity(_ context.Context, _ int64, _, _ float64) (bool, error) {
	return true, nil
}

// haversineDistance calculates the distance in meters between two lat/lon points.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusM = 6371000.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusM * c
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}
