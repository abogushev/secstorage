create table users(
  id uuid primary key,
  login varchar(20) unique not null,
  password varchar not null
);