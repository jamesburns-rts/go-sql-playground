package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/pressly/goose/v3"
	"log"
	"time"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// before running, run `docker-compose -f postgres.yml up`
func main() {
	driver := "postgres"
	connectionString := "user=localuser password=supersecret dbname=testdb sslmode=disable host=localhost"
	ctx := context.Background()

	migrateWithGoose(driver, connectionString)

	// test methods
	customSamples := GetAllWithCustom(ctx, driver, connectionString)
	printSamples("custom", customSamples)

	sqlxSamples := GetAllWithSqlx(ctx, driver, connectionString)
	printSamples("sqlx", sqlxSamples)

	gormSamples := GetAllWithGorm(connectionString)
	printSamples("gorm", gormSamples)
}

func printSamples(source string, samples any) {
	fmt.Printf("Samples from %s:\n", source)
	b, _ := json.MarshalIndent(samples, "", "  ")
	fmt.Println(string(b))
}

func migrateWithGoose(driverName, dataSourceName string) {
	d, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	// try to connect a few times - this is mainly for docker-compose
	for i := 0; i < 3; i++ {
		if err = d.Ping(); err == nil {
			break
		}
		log.Println("Waiting one second for db to start...")
		time.Sleep(time.Second)
	}
	if err != nil {
		_ = d.Close()
		panic(fmt.Errorf("unable to ping db: %w", err))
	}
	log.Println("DB connection successful")

	// do migrations
	_ = goose.SetDialect("postgres")
	goose.SetBaseFS(embedMigrations)
	if err := goose.Up(d, "migrations", goose.WithAllowMissing()); err != nil {
		panic(err)
	}
}
