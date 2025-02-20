package application

import (
	"first-little-server/order"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (app *App) LoadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", app.LoadOrderRoutes)

	app.router = router
}

func (app *App) LoadOrderRoutes(router chi.Router) {
	orderHandler := &order.Handler{
		Repo: &order.RedisRepo{
			Client: app.ds.getRedisClient(),
		},
	}

	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateByID)
	router.Delete("/{id}", orderHandler.DeleteByID)
}
