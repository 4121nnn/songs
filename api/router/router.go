package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
	"songs/api/resource/song"

	"songs/api/resource/health"
	"songs/api/router/middleware"
	"songs/api/router/middleware/requestlog"
	_ "songs/docs"
)

func New(l *zerolog.Logger, v *validator.Validate, db *gorm.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", health.Read)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.ContentTypeJSON)

		songAPI := song.New(l, v, db)
		r.Method("GET", "/", requestlog.NewHandler(songAPI.List, l))
		r.Method("GET", "/{id}", requestlog.NewHandler(songAPI.Read, l))
		r.Method("POST", "/", requestlog.NewHandler(songAPI.Create, l))
		r.Method("PUT", "/{id}", requestlog.NewHandler(songAPI.Update, l))
		r.Method("DELETE", "/{id}", requestlog.NewHandler(songAPI.Delete, l))
		r.Method("GET", "/info", requestlog.NewHandler(songAPI.Info, l))

	})

	return r
}
