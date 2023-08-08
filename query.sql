
-- name: CreateSampleNoReturn :exec
insert into test.sample_table (name, description, int_example)
values ($1, $2, $3);

-- name: GetAllSamples :many
select * from test.sample_table;

-- name: CreateSampleWithReturn :one
insert into test.sample_table (name, description, int_example)
values ($1, $2, $3) returning *;
