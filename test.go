package main

import (
	"fmt"
	"log"
	"sync"
)

func main() {

	buf := []byte{115, 116, 114 ,101, 97, 109, 32, 114, 101, 115, 101, 116}
	fmt.Println(">>>>>>>>>>>",string(buf))
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("aaaaaaaaa")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("bbbbbbbbbb")
	}()

	wg.Wait()
}
