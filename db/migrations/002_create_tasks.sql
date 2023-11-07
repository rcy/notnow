create table tasks(
  id uuid primary key default uuid_generate_v4(),
  created_at timestamptz not null default now(),
  user_id uuid references users not null,
  summary text not null
);
---- create above / drop below ----
drop table tasks;
