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
