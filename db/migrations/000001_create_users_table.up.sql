CREATE TABLE IF NOT EXISTS users(
  id serial PRIMARY KEY,
  username VARCHAR (50) UNIQUE NOT NULL,
  password VARCHAR (64) NOT NULL,
  email VARCHAR (300) UNIQUE NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW()
);