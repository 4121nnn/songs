package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"net/http"
	"songs/api/router"
	"songs/config"
	"songs/util/logger"
	"songs/util/validator"
	"strconv"
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

	db, err := setupDatabase(c)
	if err != nil {
		l.Fatal().Err(err).Msg("DB connection setup failure")
		return
	}

	if c.DB.AutoMigrate {
		if err := runMigrations(c, l); err != nil {
			l.Fatal().Err(err).Msg("Migrations setup failure")
			return
		}
	}

	r := router.New(l, v, db)

	handler := setupCors(c, r)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Server.Port),
		Handler:      handler,
		ReadTimeout:  c.Server.TimeoutRead,
		WriteTimeout: c.Server.TimeoutWrite,
		IdleTimeout:  c.Server.TimeoutIdle,
	}

	l.Info().Msgf("Sever started %v", s.Addr)
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Fatal().Err(err).Msg("Server startup failure")
	}
}

func setupDatabase(c *config.Conf) (*gorm.DB, error) {
	dbString := fmt.Sprintf(fmtDBString, c.DB.Host, c.DB.Username, c.DB.Password, c.DB.DBName, c.DB.Port)
	var logLevel gormlogger.LogLevel
	if c.DB.Debug {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	db, err := gorm.Open(gormPostgres.Open(dbString), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(c *config.Conf, l *zerolog.Logger) error {
	dsn := fmt.Sprintf(fmtDBString, c.DB.Host, c.DB.Username, c.DB.Password, c.DB.DBName, c.DB.Port)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer conn.Close()

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	l.Info().Msg("Migrations applied successfully")
	return nil
}

func setupCors(c *config.Conf, r http.Handler) http.Handler {
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

	return corsHandler.Handler(r)
}
