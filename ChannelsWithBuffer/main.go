package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Point2D struct {
	x int
	y int
}

const numberOfThreads = 8

var (
	r  = regexp.MustCompile(`\((\d*),(\d*)\)`)
	wg = sync.WaitGroup{}
)

func findArea(inputChannel chan string) {
	for pointStr := range inputChannel {
		var points []Point2D

		for _, p := range r.FindAllStringSubmatch(pointStr, -1) {
			x, _ := strconv.Atoi(p[1])
			y, _ := strconv.Atoi(p[2])
			points = append(points, Point2D{x, y})
		}

		area := 0.0
		for i := 0; i < len(points); i++ {
			a, b := points[i], points[(i+1)%len(points)]

			area += float64(a.x*b.y) - float64(a.y*b.x)
		}
	}
	wg.Done()
}

func main() {
	absPath, _ := filepath.Abs("./")
	data, _ := ioutil.ReadFile(filepath.Join(absPath, "polygons.txt"))
	text := string(data)

	inputChannel := make(chan string, 1000)
	for i := 0; i < numberOfThreads; i++ {
		go findArea(inputChannel)
	}
	wg.Add(numberOfThreads)
	start := time.Now()
	for _, line := range strings.Split(text, "\n") {
		inputChannel <- line
	}
	close(inputChannel)
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Processing took %v \n", elapsed)
}
