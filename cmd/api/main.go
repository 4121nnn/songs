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

	l.Info().Msg("Starting Songs Library server")

	db, err := setupDatabase(c, l)
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

	handler := setupCors(c, r, l)

	l.Info().Msgf("Starting server on port %d", c.Server.Port)
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Server.Port),
		Handler:      handler,
		ReadTimeout:  c.Server.TimeoutRead,
		WriteTimeout: c.Server.TimeoutWrite,
		IdleTimeout:  c.Server.TimeoutIdle,
	}

	l.Info().Msgf("Server started at %s", s.Addr)
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Fatal().Err(err).Msg("Server startup failure")
	}
}

func setupDatabase(c *config.Conf, l *zerolog.Logger) (*gorm.DB, error) {
	l.Debug().Msg("Setting up database connection...")
	dbString := fmt.Sprintf(fmtDBString, c.DB.Host, c.DB.Username, c.DB.Password, c.DB.DBName, c.DB.Port)
	l.Debug().Str("DB Connection String", dbString).Msg("Database connection string formatted")

	var logLevel gormlogger.LogLevel
	if c.DB.Debug {
		logLevel = gormlogger.Info
		l.Debug().Msg("Database debug mode is ON")
	} else {
		logLevel = gormlogger.Error
		l.Debug().Msg("Database debug mode is OFF")
	}

	db, err := gorm.Open(gormPostgres.Open(dbString), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		l.Error().Err(err).Msg("Failed to connect to the database")
		return nil, err
	}

	l.Info().Msg("Database connection successfully established")
	return db, nil
}

func runMigrations(c *config.Conf, l *zerolog.Logger) error {
	dsn := fmt.Sprintf(fmtDBString, c.DB.Host, c.DB.Username, c.DB.Password, c.DB.DBName, c.DB.Port)
	l.Debug().Str("dsn", "****").Msg("Connecting to the database for migration")

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

	l.Info().Msg("Running database migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	l.Info().Msg("Migrations applied successfully")
	return nil
}

func setupCors(c *config.Conf, r http.Handler, l *zerolog.Logger) http.Handler {
	var origin string
	if c.FR.Port == 80 {
		origin = fmt.Sprintf("http://%s", c.FR.Host)
	} else {
		origin = fmt.Sprintf("http://%s:%s", c.FR.Host, strconv.Itoa(c.FR.Port))
	}

	l.Debug().Str("origin", origin).Msg("Setting up CORS for the origin")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	})

	l.Info().Str("origin", origin).Msg("CORS setup completed successfully")

	return corsHandler.Handler(r)
}
