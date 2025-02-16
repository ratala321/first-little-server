package main

import (
	"fmt"
	"net/http"
)

func main() {
	server := &http.Server{
		Addr:    ":3000",
		Handler: http.HandlerFunc(EntryHandler),
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Root error:", err)
		return
	}
}

// EntryHandler is the entry point for all http requests.
// It acts as the controller.
func EntryHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		break
	case "POST":
		break
	default:
		break
	}
}
