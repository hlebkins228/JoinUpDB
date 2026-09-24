-- +goose Up
create table if not exists image (
    id serial primary key,
    name text not null,
    data bytea not null
);

create table if not exists location (
    id serial primary key,
    name text not null,
    longitude float not null,
    latitude float not null,
    address text not null
);

create table if not exists users (
    id serial primary key,
    name text not null,
    age int not null check (age > 0),
    login text not null,
    password text not null,
    created_at timestamp not null,
    city text not null,
    telegram_login text,
    avatar_id int references image(id)
);

create table if not exists event (
    id serial primary key,
    name text not null,
    description text,
    created_at timestamp not null,
    updated_at timestamp not null,
    event_date timestamp not null,
    telegram_chat_url text,
    city text not null,
    creator_id int not null references users(id) on delete cascade,
    location_id int not null references location(id),
    image_id int references image(id),
    deleted boolean not null default false
);

create table if not exists member (
    id serial primary key,
    user_id int not null references users(id) on delete cascade,
    event_id int not null references event(id) on delete cascade,
    role text not null
);

create table if not exists category (
    id serial primary key,
    subcategory_id int references category(id),
    name text not null constraint category_name_unique unique
);

create table if not exists event_category (
    event_id int not null references event(id) on delete cascade,
    category_id int not null references category(id),
    primary key (event_id, category_id)
);

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

-- +goose Down
drop table if exists subscribe;
drop table if exists preset;
drop table if exists event_category;
drop table if exists category;
drop table if exists member;
drop table if exists event;
drop table if exists users;
drop table if exists location;
drop table if exists image;
