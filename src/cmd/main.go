package main

import (
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"

	"webinar2/src/internal/controller"
	"webinar2/src/internal/repository"
	"webinar2/src/internal/service/database"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"webinar2/src/internal/config"
)

func main() {
	cfg := config.NewConfig()
	err := cfg.ReadConfigFromFile("./config.json")
	if err != nil {
		log.Fatalf("reading configuration %+v", err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	db, err := database.NewDatabase(cfg.DbDSN, logger)
	if err != nil {
		log.Fatalf("database connection %+v", err)
	}

	// так делать нельзя, но я сделаю для упрощения примера
	migrations(db.DB(), cfg.MigrationPath)

	repo := repository.NewBaseRepository(db.DB(), logger)

	controller := controller.NewBaseController(logger, repo)

	r := chi.NewRouter()
	r.Mount("/", controller.Route())

	log.Println("Server start ... ")
	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatalf("server error: %+v", err)
	}
}

func migrations(db *sqlx.DB, path string) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("mirations: %+v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		path,
		"postgres", driver)
	if err != nil {
		log.Fatalf("read migrations: %+v", err)
	}
	m.Up()
}
