-- +goose Up
create index if not exists event_search_city_date_idx
    on event (city, event_date, id)
    where deleted = false;

-- +goose Down
drop index if exists event_search_city_date_idx;
