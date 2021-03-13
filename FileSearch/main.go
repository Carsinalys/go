package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

var (
	matches   []string
	waitGroup = sync.WaitGroup{}
	lock      = sync.Mutex{}
)

func fileSearch(root string, fileName string) {
	fmt.Println("Searching in", root)
	files, _ := ioutil.ReadDir(root)
	for _, file := range files {
		if strings.Contains(file.Name(), fileName) {
			lock.Lock()
			matches = append(matches, filepath.Join(root, file.Name()))
			lock.Unlock()
		}

		if file.IsDir() {
			waitGroup.Add(1)
			go fileSearch(filepath.Join(root, file.Name()), fileName)
		}
	}
	waitGroup.Done()
}

func main() {
	waitGroup.Add(1)
	go fileSearch("/Users/cardinalys/Documents", "README.md")
	waitGroup.Wait()
	for _, file := range matches {
		fmt.Println("Matched", file)
	}
}
