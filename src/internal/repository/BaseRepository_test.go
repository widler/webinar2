package repository_test

import (
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"webinar2/src/internal/config"
	"webinar2/src/internal/repository"
	"webinar2/src/internal/service/database"
)

func TestBaseRepositorySuite(t *testing.T) {
	suite.Run(t, new(TestBaseRepositoryStaticSuite))
}

type log interface {
	Info(args ...interface{})
}

func InitDb(dsn string, logger log, migrationsPath string) *database.Database {
	db, err := database.NewDatabase(dsn, logger)
	if err != nil {
		panic("database error")
	}
	migrations(db.DB(), migrationsPath)
	return db
}

func migrations(db *sqlx.DB, path string) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		fmt.Println(err)
		panic("mirations")
	}
	m, err := migrate.NewWithDatabaseInstance(
		path,
		"postgres", driver)
	if err != nil {
		fmt.Println(err)
		panic("read migrations")
	}
	m.Up()
}

type TestBaseRepositoryStaticSuite struct {
	suite.Suite

	db     *database.Database
	logger log
	repo   *repository.BaseRepository
}

func (s *TestBaseRepositoryStaticSuite) SetupSuite() {
	cfg := config.NewConfig()
	err := cfg.ReadConfigFromFile("/home/denis/GolandProjects/webinar2/config-test.json")
	if err != nil {
		panic(err)
	}

	s.logger = logrus.New()
	db := InitDb(cfg.DbDSN, s.logger, cfg.MigrationPath)
	s.db = db
	s.repo = repository.NewBaseRepository(s.db.DB(), s.logger)
}

func (s *TestBaseRepositoryStaticSuite) SetupTest() {
	s.clearDB()
}

func (s *TestBaseRepositoryStaticSuite) clearDB() {
	_, err := s.db.DB().Exec("TRUNCATE TABLE storage")
	s.NoError(err, "очистка БД")
}

func (s *TestBaseRepositoryStaticSuite) TearDownSuite() {
	s.NoError(s.db.Close())
}

func (s *TestBaseRepositoryStaticSuite) TestBaseRepository_Get() {
	tests := []struct {
		name          string
		givenName     string
		givenValue    string
		expectedValue string
	}{
		{
			name:          "Тест на получение данных",
			givenName:     "var 1",
			givenValue:    "var 2",
			expectedValue: "var 2",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.clearDB()
			_, err := s.db.DB().Exec("INSERT INTO storage(key, value) VALUES ('" + test.givenName + "', '" + test.givenValue + "')")
			s.NoError(err, "Добавление метрики")

			str, err := s.repo.Get(test.givenName)
			s.NoError(err, "получение значения метрики")
			s.Equal(str, test.expectedValue)
		})
	}
}

func (s *TestBaseRepositoryStaticSuite) TestBaseRepository_Set() {
	tests := []struct {
		name          string
		givenName     string
		givenValue    string
		expectedValue string
	}{
		{
			name:          "Тест на сохранение данных",
			givenName:     "var 4",
			givenValue:    "var 5",
			expectedValue: "var 5",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.clearDB()

			err := s.repo.Set(test.givenName, test.givenValue)
			s.NoError(err, "Добавление метрики")

			res, err := s.db.DB().Query("SELECT value FROM storage WHERE key='" + test.givenName + "'")
			res.Next()
			var value string
			err = res.Scan(&value)
			s.NoError(err, "получение значения метрики")
			s.Equal(value, test.expectedValue)
		})
	}
}
