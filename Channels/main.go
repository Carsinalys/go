package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	year      = regexp.MustCompile(`\d{4}`)
	matchData [3]int
)

func parseToArray(textChannel chan string, arrChannel chan []string) {
	for text := range textChannel {
		lines := strings.Split(text, "\n")
		metarSlice := make([]string, 0, len(lines))
		for _, line := range lines {
			if res := year.FindAllString(line, -1); len(res) > 0 {
				metarSlice = append(metarSlice, res[0])
			}
		}
		arrChannel <- metarSlice
	}
	close(arrChannel)
}

func mineData(arrChannel chan []string, matchChannel chan [3]int) {
	for arr := range arrChannel {
		for _, v := range arr {
			if v == "2020" {
				matchData[0]++
			}

			if v == "2021" {
				matchData[1]++
			}

			if v == "2022" {
				matchData[2]++
			}
		}
	}
	matchChannel <- matchData
	close(matchChannel)
}

func main() {
	textChannel := make(chan string)
	arrChannel := make(chan []string)
	matchChannel := make(chan [3]int)

	go parseToArray(textChannel, arrChannel)

	go mineData(arrChannel, matchChannel)

	absPath, _ := filepath.Abs("./data")
	files, _ := ioutil.ReadDir(absPath)
	start := time.Now()
	for _, file := range files {
		data, err := ioutil.ReadFile(filepath.Join(absPath, file.Name()))
		if err != nil {
			panic(err)
		}
		text := string(data)
		textChannel <- text
	}
	close(textChannel)
	result := <-matchChannel
	elapsed := time.Since(start)
	fmt.Printf("%v\n", result)
	fmt.Printf("Processing took%s\n", elapsed)
}
