-- +goose Up
create table if not exists event_deletion_log (
    id bigserial primary key,
    event_id int not null references event(id) on delete cascade,
    deleted_at timestamp not null default (now() at time zone 'utc')
);

-- +goose StatementBegin
create or replace function log_event_deletion()
returns trigger
language plpgsql
as $$
begin
    if old.deleted = false and new.deleted = true then
        insert into event_deletion_log (event_id)
        values (new.id);
    end if;

    return new;
end;
$$;
-- +goose StatementEnd

create trigger event_deletion_log_trigger
after update of deleted on event
for each row
execute function log_event_deletion();

-- +goose Down
drop trigger if exists event_deletion_log_trigger on event;
drop function if exists log_event_deletion();
drop table if exists event_deletion_log;
