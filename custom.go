package main

import (
	"context"
	"database/sql"
	"log"
	"time"
)

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
