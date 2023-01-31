package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

type CustomSample struct {
	ID          int
	Name        string
	Description sql.NullString
	IntExample  sql.NullInt64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

func GetAllWithCustom(ctx context.Context, driverName, dataSourceName string) []CustomSample {
	// connect to db
	d, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	// query
	rows, err := d.QueryContext(ctx, "select * from test.sample_table")
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal
	samples := make([]CustomSample, 0)
	for rows.Next() {
		var c CustomSample
		err = rows.Scan(&c.ID, &c.Name, &c.Description, &c.IntExample, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
		if err != nil {
			log.Fatal(err)
		}
		samples = append(samples, c)
	}
	return samples
}

type SqlxSample struct {
	ID          int            `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	IntExample  sql.NullInt64  `db:"int_example"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

func GetAllWithSqlx(ctx context.Context, driverName, dataSourceName string) []SqlxSample {
	db, err := sqlx.ConnectContext(ctx, driverName, dataSourceName)
	if err != nil {
		log.Fatalln(err)
	}

	samples := make([]SqlxSample, 0)
	err = db.SelectContext(ctx, &samples, "select * from test.sample_table")
	if err != nil {
		panic(err)
	}
	return samples
}

type SampleTable struct {
	gorm.Model
	ID          int
	Name        string
	Description *string
	IntExample  *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func GetAllWithGorm(dataSourceName string) []SampleTable {
	conn := postgres.Open(dataSourceName)
	gd, err := gorm.Open(conn, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "test.",
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatal(fmt.Errorf("creating migrate: %w", err))
	}

	samples := make([]SampleTable, 0)
	_ = gd.Find(&samples)
	return samples
}
