package boid

import (
	"math"
)

// Vector2d - simple implementation point on screen
type Vector2d struct {
	X float64
	Y float64
}

// Add - add coords to existing point
func (v1 *Vector2d) Add(v2 Vector2d) Vector2d {
	return Vector2d{v1.X + v2.X, v1.Y + v2.Y}
}

// Subtract - subtract coords from existing point
func (v1 *Vector2d) Subtract(v2 Vector2d) Vector2d {
	return Vector2d{v1.X - v2.X, v1.Y - v2.Y}
}

// Multiply - multiply coords existing point
func (v1 *Vector2d) Multiply(v2 Vector2d) Vector2d {
	return Vector2d{v1.X * v2.X, v1.Y * v2.Y}
}

// AddV - add float to coords existing point
func (v1 *Vector2d) AddV(d float64) Vector2d {
	return Vector2d{v1.X + d, v1.Y + d}
}

// DivisionV - divide by float coords existing point
func (v1 *Vector2d) DivisionV(d float64) Vector2d {
	return Vector2d{v1.X / d, v1.Y / d}
}

// MultiplyV - multiply by float coords existing point
func (v1 *Vector2d) MultiplyV(d float64) Vector2d {
	return Vector2d{v1.X * d, v1.Y * d}
}

//Limit - calc borders for point
func (v1 *Vector2d) Limit(lower, upper float64) Vector2d {
	return Vector2d{math.Min(math.Max(v1.X, lower), upper), math.Min(math.Max(v1.Y, lower), upper)}
}

// Distance - return distance between two points
func (v1 *Vector2d) Distance(v2 Vector2d) float64 {
	return math.Sqrt(math.Pow(v1.X-v2.X, 2) + math.Pow(v1.Y-v2.Y, 2))
}
