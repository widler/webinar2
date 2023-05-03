package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	DbDSN         string `json:"db_dsn"`
	ServerAddress string `json:"server_address"`
	MigrationPath string `json:"migration_path"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ReadConfigFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	err = json.Unmarshal(bytes, c)
	if err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}
	return nil
}

func (c Config) DSN() string {
	return c.DbDSN
}
