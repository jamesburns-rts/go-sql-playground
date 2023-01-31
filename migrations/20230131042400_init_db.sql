-- +goose Up
create schema test;

create table test.sample_table
(
    id   serial                        not null primary key,
    name text not null,
    description text,
    int_example int,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted_at timestamp
);

insert into test.sample_table (name, description, int_example) values ('first', 'with description', 1);
insert into test.sample_table (name) values ('just name');

-- +goose Down
drop table test.sample_table;
drop schema test;