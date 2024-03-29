// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: query.sql

package sqlcdb

import (
	"context"
)

const createSampleNoReturn = `-- name: CreateSampleNoReturn :exec
insert into test.sample_table (name, description, int_example)
values ($1, $2, $3)
`

type CreateSampleNoReturnParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IntExample  *int32  `json:"intExample"`
}

func (q *Queries) CreateSampleNoReturn(ctx context.Context, arg CreateSampleNoReturnParams) error {
	_, err := q.db.Exec(ctx, createSampleNoReturn, arg.Name, arg.Description, arg.IntExample)
	return err
}

const createSampleWithReturn = `-- name: CreateSampleWithReturn :one
insert into test.sample_table (name, description, int_example)
values ($1, $2, $3) returning id, name, description, int_example, created_at, updated_at, deleted_at
`

type CreateSampleWithReturnParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IntExample  *int32  `json:"intExample"`
}

func (q *Queries) CreateSampleWithReturn(ctx context.Context, arg CreateSampleWithReturnParams) (TestSampleTable, error) {
	row := q.db.QueryRow(ctx, createSampleWithReturn, arg.Name, arg.Description, arg.IntExample)
	var i TestSampleTable
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IntExample,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getAllSamples = `-- name: GetAllSamples :many
select id, name, description, int_example, created_at, updated_at, deleted_at from test.sample_table
`

func (q *Queries) GetAllSamples(ctx context.Context) ([]TestSampleTable, error) {
	rows, err := q.db.Query(ctx, getAllSamples)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TestSampleTable{}
	for rows.Next() {
		var i TestSampleTable
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.IntExample,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
