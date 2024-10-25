package helper

import (
	"math"
	"shareway/schemas"

	"github.com/twpayne/go-polyline"
)

const (
	earthRadius = 6371 // Earth radius in kilometers
	minDistance = 0.1  // Minimum distance in kilometers
)

func haversineDistance(p1, p2 schemas.Point) float64 {
	dLat := toRadians(p2.Lat - p1.Lat)
	dLon := toRadians(p2.Lng - p1.Lng)
	lat1 := toRadians(p1.Lat)
	lat2 := toRadians(p2.Lat)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func OptimizeRoutePoints(points []schemas.Point) []schemas.Point {
	if len(points) < 2 {
		return points
	}

	optimized := []schemas.Point{points[0]}

	for i := 1; i < len(points); i++ {
		if haversineDistance(optimized[len(optimized)-1], points[i]) >= minDistance {
			optimized = append(optimized, points[i])
		}
	}

	// Always include the last point
	if optimized[len(optimized)-1] != points[len(points)-1] {
		optimized = append(optimized, points[len(points)-1])
	}

	return optimized
}

func DecodePolyline(encodedPolyline string) []schemas.Point {
	buf := []byte(encodedPolyline)
	coords, _, _ := polyline.DecodeCoords(buf)

	var points []schemas.Point
	// the coords is a slice of slices, each slice contains two elements: latitude and longitude
	// and have order to create a route on the map the previous point use for encode the next point
	for _, coord := range coords {
		points = append(points, schemas.Point{Lat: coord[0], Lng: coord[1]})
	}

	return points
}
