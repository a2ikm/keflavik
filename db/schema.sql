CREATE TABLE users (
  id serial PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL
);
