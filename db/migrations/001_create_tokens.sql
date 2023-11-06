create extension if not exists "uuid-ossp";

create table tokens (
  id serial primary key,
  token jsonb not null,
  uuid uuid not null default uuid_generate_v4()
);
---- create above / drop below ----
drop table tokens;
