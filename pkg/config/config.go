package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Postgres struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		Schema   string `yaml:"schema"`
		SSLMode  string `yaml:"sslmode"`
		MaxConns int    `yaml:"max_conns"`
		MinConns int    `yaml:"min_conns"`
	} `yaml:"postgres"`
	Migration struct {
		Dir string `yaml:"dir"`
	} `yaml:"migration"`
}

func Load() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read congig: %s", err)
	}

	return &cfg
}

// DSN returns PostgreSQL connection string (DSN) from config
func (c *Config) DSN() string {
	pg := c.Postgres
	sslmode := pg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	port := pg.Port
	if port == 0 {
		port = 5432
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pg.Host, port, pg.User, pg.Password, pg.DBName, sslmode,
	)
}
