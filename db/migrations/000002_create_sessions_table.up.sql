CREATE TABLE IF NOT EXISTS sessions(
  id serial PRIMARY KEY,
  user_id INT NOT NULL,
  token VARCHAR (64) NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_user
    FOREIGN KEY(user_id) 
	    REFERENCES users(id)
);