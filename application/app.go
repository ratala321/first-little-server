package application

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type App struct {
	router http.Handler
	ds     *Datastore
	config Config
}

func NewApp(ctx context.Context, config Config) *App {
	app := &App{
		ds:     NewDatastore(ctx, config),
		config: config,
	}

	app.LoadRoutes()

	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.ServerPort),
		Handler: app.router,
	}

	if err := app.ds.Ping(ctx); err != nil {
		return err
	}

	// Wrap defer call in anonymous function because defer keyword does not work when returning an error.
	defer func() {
		if err := app.ds.Close(ctx); err != nil {
			fmt.Println("failed to close datastore", err)
		}
	}()

	fmt.Println("Starting server")

	channel := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()

		if err != nil {
			channel <- fmt.Errorf("failed to start server: %w", err)
		}

		close(channel)
	}()

	select {
	case err := <-channel:
		return err
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer shutdownCancel()

		return server.Shutdown(shutdownCtx)
	}
}
