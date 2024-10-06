package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"songs/api/router"
	"songs/config"
	"songs/util/logger"
	"songs/util/validator"
	"strconv"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/cors"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

//	@title			Songs Library
//	@version		1.0
//	@description	This is test project for Effective Mobile
//
// @host		localhost:8080
// @basePath	/v1
func main() {
	c := config.New()
	l := logger.New(c.Server.Debug)
	v := validator.New()

	var logLevel gormlogger.LogLevel
	if c.DB.Debug {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	dbString := fmt.Sprintf(fmtDBString, c.DB.Host, c.DB.Username, c.DB.Password, c.DB.DBName, c.DB.Port)
	db, err := gorm.Open(gormPostgres.Open(dbString), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		l.Fatal().Err(err).Msg("DB connection start failure")
		return
	}

	r := router.New(l, v, db)

	// Set up CORS options
	var origin string
	if c.FR.Port == 80 {
		origin = fmt.Sprintf("http://%s", c.FR.Host)
	} else {
		origin = fmt.Sprintf("http://%s:%s", c.FR.Host, strconv.Itoa(c.FR.Port))
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	})

	// Wrap the router with CORS middleware
	handler := corsHandler.Handler(r)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Server.Port),
		Handler:      handler,
		ReadTimeout:  c.Server.TimeoutRead,
		WriteTimeout: c.Server.TimeoutWrite,
		IdleTimeout:  c.Server.TimeoutIdle,
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		l.Info().Msgf("Shutting down server %v", s.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), c.Server.TimeoutIdle)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			l.Error().Err(err).Msg("Server shutdown failure")
		}

		sqlDB, err := db.DB()
		if err == nil {
			if err = sqlDB.Close(); err != nil {
				l.Error().Err(err).Msg("DB connection closing failure")
			}
		}

		close(closed)
	}()

	l.Info().Msgf("Starting server %v", s.Addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Fatal().Err(err).Msg("Server startup failure")
	}

	<-closed
	l.Info().Msgf("Server shutdown successfully")
}
