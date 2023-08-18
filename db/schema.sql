CREATE TABLE users (
  id serial PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE sessions (
  id serial PRIMARY KEY,
  user_id integer NOT NULL,
  access_token VARCHAR(50) UNIQUE NOT NULL
);
