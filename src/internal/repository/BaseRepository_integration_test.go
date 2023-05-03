package repository

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"webinar2/src/internal/controller"
	"webinar2/src/internal/service/database"
)

const (
	defaultPostgresImageTag = "postgres:latest"
	defaultPort             = 5432
	defaultUser             = "postgres"
	defaultPassword         = "postgres"
	defaultDbName           = "test"
	defaultMigrationsPath   = "../../../migrations"
)

type testLogger struct {
	t *testing.T
}

func (r testLogger) Info(args ...interface{}) {
	r.t.Log(args...)
}

type PostgresContainerSuite struct {
	suite.Suite
	container testcontainers.Container
	repo      controller.Storage
}

// TestBaseRepository_PostgresContainerSuite wrapper for run the suite as normal test
func TestBaseRepository_PostgresContainerSuite(t *testing.T) {
	suite.Run(t, new(PostgresContainerSuite))
}

// SetupSuite will run before each test and is responsible to set up db container and db itself
func (r *PostgresContainerSuite) SetupSuite() {
	// setup postgres docker container
	if err := r.setupPostgresContainer(); err != nil {
		// skip test if something goes wrong at this stage
		r.T().Skipf("container setup is failed, cause: %s", err)
	}
	// check our struct has initialized container field
	require.NotNil(r.T(), r.container)
	// get container host
	host, err := r.container.Host(context.Background())
	require.NoError(r.T(), err)
	// get container port
	port, err := r.container.MappedPort(context.Background(), nat.Port(strconv.Itoa(defaultPort)))
	require.NoError(r.T(), err)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", defaultUser, defaultPassword, host, port.Int(), defaultDbName)
	// logger stub
	logger := &testLogger{t: r.T()}
	// init db
	db, err := database.NewDatabase(dsn, logger)
	require.NoError(r.T(), err)
	driver, err := postgres.WithInstance(db.DB().DB, &postgres.Config{})
	require.NoError(r.T(), err)

	migrationsAbsPath, err := filepath.Abs(defaultMigrationsPath)
	require.NoError(r.T(), err)
	migrationsAbsPath = filepath.Join("file://", migrationsAbsPath)
	m, err := migrate.NewWithDatabaseInstance(migrationsAbsPath, "postgres", driver)
	require.NoError(r.T(), err)
	// run migrations
	require.NoError(r.T(), m.Up())
	// init repository
	r.repo = NewBaseRepository(db.DB(), logger)
}

// TearDownSuite will run after each test and is responsible to terminate our container
func (r *PostgresContainerSuite) TearDownSuite() {
	if err := r.container.Terminate(context.Background()); err != nil {
		r.T().Skipf("container tearDown is failed, cause: %s", err)
	}
}

// setupPostgresContainer setup and run container
func (r *PostgresContainerSuite) setupPostgresContainer() error {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{},
		Image:          defaultPostgresImageTag,
		WaitingFor:     wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
		ExposedPorts:   []string{strconv.Itoa(defaultPort)},
		Env: map[string]string{
			"POSTGRES_USER":     defaultUser,
			"POSTGRES_PASSWORD": defaultPassword,
			"POSTGRES_DB":       defaultDbName,
		},
	}
	err := req.Validate()
	require.NoError(r.T(), err)
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}
	r.container = container
	return nil
}

// TestBaseRepository_GetSet this is our test - for simplicity reasons all cases are modeled as subtests
func (r *PostgresContainerSuite) TestBaseRepository_GetSet() {

	key := "key_1"
	val := "val_1"

	r.T().Run("try to get non existing value", func(t *testing.T) {
		get, err := r.repo.Get(key)
		require.Equal(t, sql.ErrNoRows, err)
		require.Equal(t, "", get)
	})

	r.T().Run("set value", func(t *testing.T) {
		err := r.repo.Set(key, val)
		require.NoError(t, err)
	})

	r.T().Run("get value", func(t *testing.T) {
		get, err := r.repo.Get(key)
		require.NoError(t, err)
		require.Equal(t, val, get)
	})

}
