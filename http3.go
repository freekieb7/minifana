package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		requestMessage := make([]byte, 1024)
		request.Body.Read(requestMessage)
		print(string(requestMessage))

	})
	http.ListenAndServe("127.0.0.1:36895", nil)
}
