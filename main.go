package main

import (
	"fmt"
	"log"
	"net/http"
	"progressive-cache/cache"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var request cache.Request

		if err := request.ReadFrom(r); err != nil {
			log.Println(err)
			return
		}

		if response, found := cache.Get(request); found {
			log.Println("From cache")
			response.Proxy(w)
			return
		}

		response, err := cache.SendHttpRequest(request)
		if err != nil {
			log.Println(err)
			return
		}
		cache.Add(request, response)
		response.Proxy(w)
	})

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(":80", nil))
}