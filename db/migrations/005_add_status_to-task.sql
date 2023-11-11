alter table tasks add column status text not null default 'open';
---- create above / drop below ----
alter table tasks drop column status;
