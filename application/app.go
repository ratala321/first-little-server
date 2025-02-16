package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
	router http.Handler
}

func NewApp() *App {
	return &App{
		router: LoadRoutes(),
	}
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	err := server.ListenAndServe()

	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
