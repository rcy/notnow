create table task_events(
  id uuid primary key default uuid_generate_v4(),
  created_at timestamptz not null default now(),
  user_id uuid references users not null,
  task_id uuid references tasks not null,
  event_id text not null
);
---- create above / drop below ----
drop table task_events;
