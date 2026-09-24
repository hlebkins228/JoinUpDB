-- +goose Up
alter table users
    add constraint users_login_unique unique (login),
    add column role text not null default 'user';

-- +goose Down
alter table users
    drop constraint if exists users_login_unique,
    drop column if exists role;

    
