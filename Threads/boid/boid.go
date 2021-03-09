package boid

import (
	"math"
	"math/rand"
	"time"
)

//Boid - base boid struct
type Boid struct {
	Position Vector2d
	Velocity Vector2d
	id       int
}

const (
	//Screenwidth screen width
	Screenwidth = 640
	//Screenheight screen height
	Screenheight = 360
	//BoidCount number of boids on screen
	BoidCount  = 500
	viewRadius = 13
	adjRate    = 0.015
)

// base maps for boids
var (
	Boids   [BoidCount]*Boid
	BoidMap [Screenwidth + 1][Screenheight + 1]int
)

func (b *Boid) calcAcceleration() Vector2d {
	upper, lower := b.Position.AddV(viewRadius), b.Position.AddV(-viewRadius)
	avgVelocity := Vector2d{0, 0}
	count := 0.0

	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, Screenwidth); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, Screenheight); j++ {
			if otherboid := BoidMap[int(i)][int(j)]; otherboid != -1 && otherboid != b.id {
				if dist := Boids[otherboid].Position.Distance(b.Position); dist < viewRadius {
					count++
					avgVelocity = avgVelocity.Add(Boids[otherboid].Velocity)
				}
			}
		}
	}

	accel := Vector2d{0, 0}
	if count > 0 {
		avgVelocity = avgVelocity.DivisionV(count)
		avgVelocity = avgVelocity.Subtract(b.Velocity)
		accel = avgVelocity.MultiplyV(adjRate)
	}

	return accel
}

func (b *Boid) moveOne() {
	b.Velocity = b.Velocity.Add(b.calcAcceleration())
	b.Velocity = b.Velocity.Limit(-1, 1)
	BoidMap[int(b.Position.X)][int(b.Position.Y)] = -1
	b.Position = b.Position.Add(b.Velocity)
	BoidMap[int(b.Position.X)][int(b.Position.Y)] = b.id
	next := b.Position.Add(b.Velocity)

	if next.X >= Screenwidth || next.X < 0 {
		b.Velocity = Vector2d{-b.Velocity.X, b.Velocity.Y}
	}

	if next.Y >= Screenheight || next.Y < 0 {
		b.Velocity = Vector2d{b.Velocity.X, -b.Velocity.Y}
	}
}

func (b *Boid) start() {
	for {
		b.moveOne()
		time.Sleep(5 * time.Millisecond)
	}
}

// CreateBoid creates single boid with random coords
func CreateBoid(bid int) {
	b := Boid{
		Position: Vector2d{rand.Float64() * Screenwidth, rand.Float64() * Screenheight},
		Velocity: Vector2d{(rand.Float64() * 2) - 1.0, (rand.Float64() * 2) - 1.0},
		id:       bid,
	}

	Boids[bid] = &b
	BoidMap[int(b.Position.X)][int(b.Position.Y)] = b.id
	go b.start()
}
