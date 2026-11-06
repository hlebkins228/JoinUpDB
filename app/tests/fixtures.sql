insert into users (name, age, login, password, created_at, city)
values ('Organizer', 25, 'organizer', 'unused', '2026-01-01', 'Moscow'),
       ('Reviewer', 24, 'reviewer', 'unused', '2026-01-01', 'Moscow');
insert into location (name, longitude, latitude, address)
values ('Park', 37.62, 55.75, 'Park entrance');
insert into event (name, description, telegram_chat_url, created_at, updated_at, event_date, city, creator_id, location_id)
values ('Past meetup', '', '', '2026-01-01', '2026-01-01', '2026-01-02', 'Moscow', 1, 1),
       ('Future meetup', '', '', '2026-01-01', '2026-01-01', '2100-01-01', 'Moscow', 1, 1);
