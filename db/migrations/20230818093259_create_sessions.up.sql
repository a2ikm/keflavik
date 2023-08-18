CREATE TABLE sessions (
  id serial PRIMARY KEY,
  user_id integer NOT NULL,
  access_token VARCHAR(255) UNIQUE NOT NULL
);
