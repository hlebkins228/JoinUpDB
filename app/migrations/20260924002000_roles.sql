-- +goose Up
-- +goose StatementBegin
do $$
begin
    if not exists (select 1 from pg_roles where rolname = 'joinup_guest') then
        create role joinup_guest login password 'guest';
    end if;
    if not exists (select 1 from pg_roles where rolname = 'joinup_user') then
        create role joinup_user login password 'user';
    end if;
    if not exists (select 1 from pg_roles where rolname = 'joinup_admin') then
        create role joinup_admin login password 'admin';
    end if;
    execute format('grant connect on database %I to joinup_guest, joinup_user, joinup_admin', current_database());
end;
$$;
-- +goose StatementEnd

grant usage on schema public to joinup_guest, joinup_user, joinup_admin;

revoke all privileges on all tables in schema public from joinup_guest, joinup_user, joinup_admin;
revoke all privileges on all sequences in schema public from joinup_guest, joinup_user, joinup_admin;

grant select on event, location, image, category, event_category, review to joinup_guest;

grant select, insert, update, delete on
    image,
    location,
    users,
    event,
    member,
    category,
    event_category,
    preset,
    subscribe,
    review,
    event_deletion_log
to joinup_user;

grant usage, select on
    image_id_seq,
    location_id_seq,
    users_id_seq,
    event_id_seq,
    member_id_seq,
    category_id_seq,
    preset_id_seq,
    subscribe_id_seq,
    review_id_seq,
    event_deletion_log_id_seq
to joinup_user;

grant all privileges on
    image,
    location,
    users,
    event,
    member,
    category,
    event_category,
    preset,
    subscribe,
    review,
    event_deletion_log
to joinup_admin;

grant all privileges on
    image_id_seq,
    location_id_seq,
    users_id_seq,
    event_id_seq,
    member_id_seq,
    category_id_seq,
    preset_id_seq,
    subscribe_id_seq,
    review_id_seq,
    event_deletion_log_id_seq
to joinup_admin;

-- +goose Down
revoke all privileges on all tables in schema public from joinup_guest, joinup_user, joinup_admin;
revoke all privileges on all sequences in schema public from joinup_guest, joinup_user, joinup_admin;
revoke usage on schema public from joinup_guest, joinup_user, joinup_admin;
-- Роли общие для кластера: при откате убираются только права в текущей БД.
-- +goose StatementBegin
do $$
begin
    execute format('revoke connect on database %I from joinup_guest, joinup_user, joinup_admin', current_database());
end;
$$;
-- +goose StatementEnd
