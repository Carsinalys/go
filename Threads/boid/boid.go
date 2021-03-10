package boid

import (
	"math"
	"math/rand"
	"sync"
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
	rwlock  = sync.RWMutex{}
)

func (b *Boid) calcAcceleration() Vector2d {
	upper, lower := b.Position.AddV(viewRadius), b.Position.AddV(-viewRadius)
	avgPosition, avgVelocity, separation := Vector2d{0, 0}, Vector2d{0, 0}, Vector2d{0, 0}
	count := 0.0

	rwlock.RLock()
	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, Screenwidth); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, Screenheight); j++ {
			if otherboid := BoidMap[int(i)][int(j)]; otherboid != -1 && otherboid != b.id {
				if dist := Boids[otherboid].Position.Distance(b.Position); dist < viewRadius {
					count++
					avgVelocity = avgVelocity.Add(Boids[otherboid].Velocity)
					avgPosition = avgPosition.Add(Boids[otherboid].Position)
					temp := b.Position.Subtract(Boids[otherboid].Position)
					temp = temp.DivisionV(dist)
					separation = separation.Add(temp)
				}
			}
		}
	}
	rwlock.RUnlock()

	accel := Vector2d{b.borderBounce(b.Position.X, Screenwidth), b.borderBounce(b.Position.Y, Screenheight)}
	if count > 0 {
		avgPosition, avgVelocity = avgPosition.DivisionV(count), avgVelocity.DivisionV(count)
		accelAlignment := avgVelocity.Subtract(b.Velocity)
		accelAlignment = accelAlignment.MultiplyV(adjRate)
		accelCohesion := avgPosition.Subtract(b.Position)
		accelCohesion = accelCohesion.MultiplyV(adjRate)
		accelSeparation := separation.MultiplyV(adjRate)
		accel = accel.Add(accelAlignment)
		accel = accel.Add(accelCohesion)
		accel = accel.Add(accelSeparation)
	}

	return accel
}

func (b *Boid) borderBounce(pos, maxBorderPos float64) float64 {
	if pos < viewRadius {
		return 1 / pos
	} else if pos > maxBorderPos-viewRadius {
		return 1 / (pos - maxBorderPos)
	}
	return 0
}

func (b *Boid) moveOne() {
	acceleration := b.calcAcceleration()
	rwlock.Lock()
	b.Velocity = b.Velocity.Add(acceleration)
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
	rwlock.Unlock()
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
