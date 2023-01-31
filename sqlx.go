package main

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

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
