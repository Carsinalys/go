package main

import (
	"image/color"
	"log"
	"threads/boid"

	"github.com/hajimehoshi/ebiten"
)

var green = color.RGBA{0, 128, 0, 1}

func update(screen *ebiten.Image) error {
	if !ebiten.IsDrawingSkipped() {
		for _, boid := range boid.Boids {
			screen.Set(int(boid.Position.X+1), int(boid.Position.Y), green)
			screen.Set(int(boid.Position.X-1), int(boid.Position.Y), green)
			screen.Set(int(boid.Position.X), int(boid.Position.Y-1), green)
			screen.Set(int(boid.Position.X), int(boid.Position.Y+1), green)
		}
	}
	return nil
}

func main() {
	for i, row := range boid.BoidMap {
		for j := range row {
			boid.BoidMap[i][j] = -1
		}
	}

	for i := 0; i < boid.BoidCount; i++ {
		boid.CreateBoid(i)
	}

	if err := ebiten.Run(update, boid.Screenwidth, boid.Screenheight, 2, "Boids in a box"); err != nil {
		log.Fatal(err)
	}
}
