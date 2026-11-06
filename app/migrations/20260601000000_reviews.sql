-- +goose Up
create table if not exists review (
    id          bigserial primary key,
    event_id    bigint    not null
        references event(id) on delete cascade,
    author_id   bigint    not null
        references users(id) on delete cascade,
    rating      integer   not null
        constraint review_rating_check check (rating between 1 and 5),
    text        text      not null
        constraint review_text_length_check check (char_length(text) <= 2000),
    created_at  timestamp not null default (now() at time zone 'utc'),
    constraint review_event_author_unique unique (event_id, author_id)
);

create index if not exists review_event_id_idx on review (event_id);
create index if not exists review_event_creator_lookup_idx on review (event_id);

-- +goose Down
drop table if exists review;
