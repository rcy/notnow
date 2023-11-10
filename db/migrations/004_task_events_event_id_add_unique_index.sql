create unique index idx_task_events_event_id on task_events (event_id);
---- create above / drop below ----
drop index if exists idx_task_events_event_id;
