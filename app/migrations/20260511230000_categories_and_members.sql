-- +goose Up
create unique index if not exists member_user_event_unique
    on member (user_id, event_id);

insert into category (name)
values
    ('спорт'),
    ('музыка'),
    ('игры'),
    ('арт'),
    ('прогулки'),
    ('учёба'),
    ('еда'),
    ('tech'),
    ('путешествия'),
    ('фото'),
    ('кино')
on conflict (name) do nothing;

-- +goose Down
drop index if exists member_user_event_unique;
