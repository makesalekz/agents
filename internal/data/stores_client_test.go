package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHaversineDistance_Within200m(t *testing.T) {
	// Two points ~100m apart in Almaty
	lat1, lon1 := 43.2380, 76.9457 // point A
	lat2, lon2 := 43.2389, 76.9457 // point B (~100m north)

	dist := haversineDistance(lat1, lon1, lat2, lon2)
	assert.Less(t, dist, 200.0, "points should be within 200m")
	assert.Greater(t, dist, 50.0, "points should be more than 50m apart")
}

func TestHaversineDistance_Outside200m(t *testing.T) {
	// Two points ~500m apart
	lat1, lon1 := 43.2380, 76.9457
	lat2, lon2 := 43.2425, 76.9457 // ~500m north

	dist := haversineDistance(lat1, lon1, lat2, lon2)
	assert.Greater(t, dist, 200.0, "points should be outside 200m")
}

func TestHaversineDistance_SamePoint(t *testing.T) {
	lat, lon := 43.2380, 76.9457
	dist := haversineDistance(lat, lon, lat, lon)
	assert.Equal(t, 0.0, dist)
}

func TestHaversineDistance_Boundary200m(t *testing.T) {
	// 200m is approximately 0.0018 degrees latitude
	lat1, lon1 := 43.2380, 76.9457
	lat2, lon2 := 43.2398, 76.9457 // ~200m north

	dist := haversineDistance(lat1, lon1, lat2, lon2)
	// Should be approximately 200m (within +/- 5m tolerance)
	assert.InDelta(t, 200.0, dist, 5.0, "should be approximately 200m")
}
