create table tokens (
  id serial primary key,
  token jsonb not null,
  session_key text not null
);
---- create above / drop below ----
drop table tokens;
