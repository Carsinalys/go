package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	money          = 100
	lock           = sync.Mutex{}
	moneyDeposited = sync.NewCond(&lock)
)

func stingy() {
	for i := 0; i < 1000; i++ {
		lock.Lock()
		money += 10
		fmt.Println("stingy sees balance of", money)
		moneyDeposited.Signal()
		lock.Unlock()
		time.Sleep(1 * time.Millisecond)
	}
	fmt.Println("stingy done")
}

func spendy() {
	for i := 0; i < 1000; i++ {
		lock.Lock()
		for money-20 < 0 {
			moneyDeposited.Wait()
		}
		money -= 20
		fmt.Println("spendy sees balance of", money)
		lock.Unlock()
		time.Sleep(1 * time.Millisecond)
	}
	fmt.Println("spendy done")
}

func main() {
	go stingy()
	go spendy()
	time.Sleep(1000 * time.Millisecond)
	print(money)
}
