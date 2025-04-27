package migrator

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	dbPostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ApplyMigrations(db *sql.DB, DB_Name string, migrationsPath string) error {
	instance, err := dbPostgres.WithInstance(db, &dbPostgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migr, err := migrate.NewWithDatabaseInstance(migrationsPath, DB_Name, instance)

	if err != nil {
		log.Fatalf("Ошибка создания экземпляра миграции: %v\n", err)
	}
	err = migr.Up()
	if err != nil && err.Error() != "no change" {
		log.Fatalf("Ошибка миграции вверх: %v\n", err)
	}
	fmt.Println("Миграции успешно применены.")

	currentVersion, dirty, err := migr.Version()
	if err != nil {
		log.Fatalf("Ошибка определения версии: %v\n", err)
	}

	fmt.Printf("Версия БД: %d, грязная: %v\n", currentVersion, dirty)

	return nil
}
