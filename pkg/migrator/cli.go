package migrator

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration CLI",
	Long:  "Run, rollback, and inspect database schema migrations.",
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all up migrations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running up migrations...")
		m, err := getMigrator()
		if err != nil {
			log.Fatalf("❌ Failed to create migrator: %v", err)
		}

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Migration up failed: %v", err)
		}

		version, dirty, _ := m.Version()
		fmt.Printf("✅ Migrated to version %d (dirty: %v)\n", version, dirty)
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback last migration",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := getMigrator()
		if err != nil {
			log.Fatalf("❌ Failed to create migrator: %v", err)
		}

		err = m.Steps(-1)
		if err != nil {
			log.Fatalf("❌ Migration down failed: %v", err)
		}

		version, dirty, _ := m.Version()
		fmt.Printf("⬅️  Rolled back to version %d (dirty: %v)\n", version, dirty)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := getMigrator()
		if err != nil {
			log.Fatalf("❌ Failed to create migrator: %v", err)
		}

		version, dirty, err := m.Version()
		if err != nil {
			fmt.Printf("⚠️  Could not get version: %v\n", err)
		} else {
			fmt.Printf("📦 Current version: %d (dirty: %v)\n", version, dirty)
		}
	},
}

// Execute запускает корневую команду
func Execute() {
	fmt.Println("Starting database migration...")
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(versionCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("❌ Failed to execute command: %v", err)
	}
}
