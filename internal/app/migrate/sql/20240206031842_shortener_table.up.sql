create table shortener
(
 uuid  uuid default gen_random_uuid() not null
  constraint shortener_pk
   primary key,
 short varchar(8)                     not null,
 url   varchar(2048)                  not null
);

