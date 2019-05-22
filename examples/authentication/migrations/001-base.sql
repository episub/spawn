CREATE EXTENSION pgcrypto;

CREATE TABLE "user" (
	user_id serial PRIMARY KEY,
	username VARCHAR(64) UNIQUE NOT NULL,
	password bytea NOT NULL,
	admin BOOLEAN NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT Now(),
	updated_at timestamptz NOT NULL DEFAULT Now()
);

ALTER TABLE public.user ADD CONSTRAINT username_lower_case CHECK (username = lower(username));

CREATE TABLE "session" (
	session_id uuid NOT NULL DEFAULT gen_random_uuid(),
	expires timestamptz NOT NULL,
	user_id INTEGER NOT NULL,
	created_at timestamptz NOT NULL DEFAULT Now(),
	updated_at timestamptz NOT NULL DEFAULT Now(),
	CONSTRAINT session_pk PRIMARY KEY (session_id),
	FOREIGN KEY (user_id) REFERENCES "user" (user_id)
);

CREATE TABLE todo (
 todo_id serial PRIMARY KEY,
 content VARCHAR NOT NULL,
 done BOOLEAN NOT NULL DEFAULT false,
 user_id INTEGER NOT NULL,
 FOREIGN KEY (user_id) REFERENCES "user" (user_id)
);

INSERT INTO "user" (user_id, username, password, admin) VALUES 
(1, 'matthew', '$2a$12$ekk6GLEiBgqYeG6AQji.5eD9lyVn5DVooN5EgFdk8I/7iC7AEsnaG', true),
(2, 'george', '$2a$12$ekk6GLEiBgqYeG6AQji.5eD9lyVn5DVooN5EgFdk8I/7iC7AEsnaG', false);

INSERT INTO todo (content, done, user_id) VALUES
('Buy milk', false, 1),
('Update documentation', false, 1),
('Fix roof', true, 1),
('Play games', true, 2),
('Rest', true, 2),
('Relax', true, 2),
('High five myself', false, 2);
