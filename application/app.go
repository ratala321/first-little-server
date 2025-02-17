package application

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func NewApp() *App {
	return &App{
		router: LoadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	redisErr := app.rdb.Ping(ctx).Err()
	if redisErr != nil {
		return fmt.Errorf("error when pinging redis: %w", redisErr)
	}

	// Wrap defer call in anonymous function because defer keyword does not work when returning an error.
	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Println("failed to close redis", err)
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
