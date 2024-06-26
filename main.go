package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go-orm-test/sqlbdb"
	"go-orm-test/sqlcdb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"io"
	"io/fs"
	"log"
	"strings"
	"time"

	//_ "github.com/lib/pq"
	"github.com/jackc/pgx/v5"
)

// TODO https://github.com/stytchauth/sqx
// TODO https://github.com/Masterminds/squirrel
// TODO https://github.com/jackc/pgx

// CustomSample to be used with built-in go sql stuff
type CustomSample struct {
	ID          int
	Name        string
	Description *string
	IntExample  *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// SqlxSample to be used with sqlx
type SqlxSample struct {
	ID          int        `db:"id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	IntExample  *int       `db:"int_example"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// SampleTable to be used with gorm - must be named after the table
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

// before running, run `docker-compose -f postgres.yml up`
// note, error handling is not done here for ease of comparison
func main() {
	driverName := "pgx"
	connectionString := "user=localuser password=supersecret dbname=testdb sslmode=disable host=localhost port=5433"
	ctx := context.Background() // you don't need to use contexts, but it's good practice

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// make connections
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// custom/sqlboiler connection
	customDBConnection, err := sql.Open(driverName, connectionString)
	if err != nil {
		panic(err)
	}
	defer safeClose(customDBConnection)

	// sqlx connection
	sqlxDBConnection, err := sqlx.ConnectContext(ctx, driverName, connectionString)
	if err != nil {
		panic(err)
	}
	defer safeClose(sqlxDBConnection)

	// gorm connection
	gormDBConnection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "test.",
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		db, _ := gormDBConnection.DB()
		_ = db.Close()
	}()

	// sqlc connection
	sqlcDBConnection, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		panic(err)
	}
	defer safeCloseCtx(sqlcDBConnection)
	sqlcQueries := sqlcdb.New(sqlcDBConnection)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// do migrations with goose
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	migrateWithGoose(customDBConnection)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// test inserts
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// custom insert
	cs := CustomSample{
		Name:        "Custom Inserted Sample",
		Description: ptr("Custom inserted description"),
	}
	_, _ = customDBConnection.ExecContext(ctx,
		"insert into test.sample_table (name, description, int_example) values ($1, $2, $3)",
		cs.Name, cs.Description, cs.IntExample,
	)

	// sqlx insert
	_, _ = sqlxDBConnection.NamedExecContext(ctx,
		"insert into test.sample_table (name, description, int_example) values (:name, :description, :int_example)",
		SqlxSample{
			Name:        "Sqlx Inserted Sample",
			Description: ptr("Sqlx inserted description"),
			IntExample:  ptr(1),
		},
	)

	// gorm insert
	_ = gormDBConnection.Create(&SampleTable{
		Name:        "Gorm Inserted Sample",
		Description: ptr("Gorm inserted description"),
	})

	// sqlc insert
	_ = sqlcQueries.CreateSampleNoReturn(ctx, sqlcdb.CreateSampleNoReturnParams{
		Name:        "SQLC Inserted Sample",
		Description: ptr("SQLC inserted description"),
		IntExample:  ptr(int32(2)),
	})

	// sqlboilere insert
	sqlbdbSample := sqlbdb.SampleTable{
		Name:        "SQLBoiler Inserted Sample",
		Description: null.StringFrom("SQLBoiler inserted description"),
		IntExample:  null.IntFrom(3),
	}
	_ = sqlbdbSample.Insert(ctx, customDBConnection, boil.Infer())

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// test selects
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// custom select
	rows, _ := customDBConnection.QueryContext(ctx, "select * from test.sample_table")
	customSamples := make([]CustomSample, 0)
	for rows.Next() {
		var c CustomSample
		_ = rows.Scan(&c.ID, &c.Name, &c.Description, &c.IntExample, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
		customSamples = append(customSamples, c)
	}
	printSamples("custom", customSamples)

	// sqlx select
	sqlxSamples := make([]SqlxSample, 0)
	_ = sqlxDBConnection.SelectContext(ctx, &sqlxSamples, "select * from test.sample_table")
	printSamples("sqlx", sqlxSamples)

	// gorm select
	gormSamples := make([]SampleTable, 0)
	_ = gormDBConnection.Find(&gormSamples)
	printSamples("gorm", gormSamples)

	// sqlc select
	sqlcSamples, _ := sqlcQueries.GetAllSamples(ctx)
	printSamples("sqlc", sqlcSamples)

	// sqlboilere insert
	sqlbSamples, _ := sqlbdb.SampleTables().All(ctx, customDBConnection)
	printSamples("sqlboiler", sqlbSamples)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// test inserts with returned
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// custom insert
	cs = CustomSample{
		Name:        "Custom Inserted Sample",
		Description: ptr("Custom inserted description"),
	}
	row := customDBConnection.QueryRowContext(ctx,
		"insert into test.sample_table (name, description, int_example) values ($1, $2, $3) returning id, created_at, updated_at",
		cs.Name, cs.Description, cs.IntExample,
	)
	_ = row.Scan(&cs.ID, &cs.CreatedAt, &cs.UpdatedAt)
	printSamples("inserted custom", cs)

	// sqlx insert
	ss := SqlxSample{
		Name:        "Sqlx Inserted Sample",
		Description: ptr("Sqlx inserted description"),
		IntExample:  ptr(1),
	}

	query, args, _ := sqlxDBConnection.BindNamed(
		"insert into test.sample_table (name, description, int_example) values (:name, :description, :int_example) returning *",
		ss,
	)
	_ = sqlxDBConnection.GetContext(ctx, &ss, query, args...)
	printSamples("sqlx inserted", ss)

	// gorm insert
	st := SampleTable{
		Name:        "Gorm Inserted Sample",
		Description: ptr("Gorm inserted description"),
	}
	_ = gormDBConnection.Create(&st)
	printSamples("gorm inserted", st)

	// sqlc insert
	sc, _ := sqlcQueries.CreateSampleWithReturn(ctx, sqlcdb.CreateSampleWithReturnParams{
		Name:        "SQLC Inserted with return Sample",
		Description: ptr("SQLC inserted description"),
		IntExample:  ptr(int32(3)),
	})
	printSamples("sqlc inserted", sc)

	// sqlboiler (handles it automatically)
	sqlbdbSample2 := sqlbdb.SampleTable{
		Name:        "SQLBoiler Inserted Sample",
		Description: null.StringFrom("SQLBoiler inserted description"),
		IntExample:  null.IntFrom(3),
	}
	_ = sqlbdbSample2.Insert(ctx, customDBConnection, boil.Infer())
	printSamples("sqlboiler inserted", sqlbdbSample2)
}

func printSamples(source string, samples any) {
	fmt.Printf("Samples from %s:\n", source)
	b, _ := json.MarshalIndent(samples, "", "  ")
	fmt.Println(string(b))
}

func migrateWithGoose(db *sql.DB) {

	// try to connect a few times - this is mainly for docker-compose
	var err error
	for i := 0; i < 3; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Println("Waiting one second for db to start...")
		time.Sleep(time.Second)
	}
	if err != nil {
		_ = db.Close()
		panic(fmt.Errorf("unable to ping db: %w", err))
	}
	log.Println("DB connection successful")

	// do migrations
	_ = goose.SetDialect("postgres")
	goose.SetBaseFS(FilteredFS{
		FS: embedMigrations,
		ShouldSkip: func(f fs.DirEntry) bool {
			return f.Name() == "2_init.sql"
		},
	})

	if err := goose.Up(db, "migrations", goose.WithAllowMissing()); err != nil {
		panic(err)
	}
}

// ptr helper function to convert any literal to a pointer
func ptr[T any](t T) *T {
	return &t
}

// safeClose is a helper method that can be used to handle closing errors - or just suppress them
func safeClose(closer io.Closer) {
	_ = closer.Close()
}

func safeCloseCtx(closer interface {
	Close(ctx2 context.Context) error
}) {
	_ = closer.Close(context.Background())
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

type FilteredFS struct {
	embed.FS
	ShouldSkip func(f fs.DirEntry) bool
}

func (f FilteredFS) ReadDir(name string) ([]fs.DirEntry, error) {
	unfiltered, err := f.FS.ReadDir(name)
	if err != nil {
		return unfiltered, err
	}
	filtered := make([]fs.DirEntry, 0, len(unfiltered))
	for _, entry := range unfiltered {
		if !f.ShouldSkip(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, err
}

func init() {
	goose.SetBaseFS(FilteredFS{
		FS: embedMigrations,
		ShouldSkip: func(f fs.DirEntry) bool {
			return strings.Contains(f.Name(), "test_data")
		},
	})
}
