create table users(
  id uuid primary key,
  login varchar(20) unique not null,
  password varchar not null
);

create table resources(
  id uuid primary key,
  user_id uuid,
  type int default 0,
  data bytea,
  meta bytea,

  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) on delete cascade
);