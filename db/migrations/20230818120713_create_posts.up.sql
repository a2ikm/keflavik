CREATE TABLE posts (
  id serial PRIMARY KEY,
  user_id integer NOT NULL,
  body text NOT NULL,
  created_at timestamp NOT NULL
);
