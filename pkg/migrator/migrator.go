package migrator

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PavelRadostev/toolkit/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// getProjectRoot returns the project root directory by finding the directory containing config/settings.yaml
func getProjectRoot() (string, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return "", fmt.Errorf("CONFIG_PATH is not set")
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for config: %w", err)
	}

	// config/settings.yaml -> remove "config/settings.yaml" to get project root
	configDir := filepath.Dir(absConfigPath)
	projectRoot := filepath.Dir(configDir) // Go up from "config" to project root

	return projectRoot, nil
}

// getMigrator creates and returns a migrate.Migrate instance
func getMigrator() (*migrate.Migrate, error) {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// Получаем корень проекта и строим абсолютный путь к миграциям
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %w", err)
	}

	migrationsDir := filepath.Join(projectRoot, cfg.Migration.Dir)
	absMigrationsDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Формируем правильный file:// URL с абсолютным путем
	migrationsURL := "file://" + absMigrationsDir

	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}
