package main

import (
	"net/http"
	"time"
)

var host = "http://localhost:8080/api/v1/ad?limit=10&offset=0&age=24&gender=F&country=TW&platform=ios"

func main() {
	request, _ := http.NewRequest(http.MethodGet, host, nil)
	_, err := http.DefaultClient.Do(request)
	if err != nil {
		println(err.Error())
		return
	}
	return
	for true {
		now := time.Now()
		times := 0
		for i := 0; i < 10000; i++ {
			s := time.Now()

			e := time.Now()
			times++
			println("request took ", e.Sub(s).Seconds(), " seconds")
		}
		println("10000 requests in ", time.Since(now).Seconds(), " seconds")
	}
}
