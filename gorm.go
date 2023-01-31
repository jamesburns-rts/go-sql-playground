package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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
