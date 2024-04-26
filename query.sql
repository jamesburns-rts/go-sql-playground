
-- name: CreateSampleNoReturn :exec
insert into test.sample_table (name, description, int_example)
values ($1, $2, $3);

-- name: GetAllSamples :many
select * from test.sample_table;

-- name: CreateSampleWithReturn :one
insert into test.sample_table (name, description, int_example)
values ($1, $2, $3) returning *;

-- name: GetDescriptions :many
select description from test.sample_table;

-- name: GetIdDescriptions :many
select id, description from test.sample_table;

-- name: GetSampleByID :one
select * from test.sample_table where id = $1;