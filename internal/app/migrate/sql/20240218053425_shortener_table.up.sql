alter table shortener
 add column user_id varchar(36) default '' not null;

create table users
(
 id         uuid        default gen_random_uuid() not null
  constraint users_pk
   primary key,
 created_at timestamptz default now()             not null
);

