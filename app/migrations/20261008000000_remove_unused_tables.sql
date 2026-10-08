-- +goose Up
drop table if exists subscribe;
drop table if exists preset;

-- +goose Down
create table if not exists preset (
    id serial primary key,
    name text not null,
    category_id int not null references category(id),
    creator_id int not null references users(id) on delete cascade
);

create table if not exists subscribe (
    id serial primary key,
    subscriber_id int not null references users(id) on delete cascade,
    user_id int not null references users(id) on delete cascade,
    subscribe_at timestamp not null
);

grant select, insert, update, delete on preset, subscribe to joinup_user;
grant usage, select on preset_id_seq, subscribe_id_seq to joinup_user;
grant all privileges on preset, subscribe to joinup_admin;
grant all privileges on preset_id_seq, subscribe_id_seq to joinup_admin;
