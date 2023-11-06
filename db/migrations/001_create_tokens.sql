create extension if not exists "uuid-ossp";

create table users(
  id uuid primary key default uuid_generate_v4(),
  created_at timestamptz not null default now(),
  email text unique not null constraint emailchk check (email != '')
);

create table sessions(
  id uuid primary key default uuid_generate_v4(),
  created_at timestamptz not null default now(),
  user_id uuid references users not null
);

create table tokens (
  id uuid primary key default uuid_generate_v4(),
  created_at timestamptz not null default now(),
  token jsonb not null,
  user_id uuid references users not null
);
---- create above / drop below ----
drop table tokens;
drop table sessions;
drop table users;
drop extension if "uuid-ossp";
