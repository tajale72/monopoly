package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	for i := 0; i < 1000; i++ {
		go func() {
			_, err := http.Get("https://wraithlike-lena-heartaching.ngrok-free.dev/roll")
			if err != nil {
				log.Println(err)
			}
		}()
	}
	time.Sleep(10 * time.Second)
}
