package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func main() {
	// A method name prefixed with new is a golang convention to indicate a constructor.
	// It is usually preferable to use these constructor methods when available.
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/hello", HelloHandler)

	server := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Root error:", err)
		return
	}
}

// HelloHandler sends "hello world" to the client.
func HelloHandler(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte("Hello World"))
}
