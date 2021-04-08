CREATE TABLE "user" (
	user_id serial PRIMARY KEY,
	username VARCHAR(64) UNIQUE NOT NULL,
	admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE todo (
 todo_id serial PRIMARY KEY,
 content VARCHAR NOT NULL,
 done BOOLEAN NOT NULL DEFAULT false,
 user_id INTEGER NOT NULL,
 FOREIGN KEY (user_id) REFERENCES "user" (user_id)
);

INSERT INTO "user" (user_id, username, admin) VALUES 
(1, 'matthew', true),
(2, 'george', false);

INSERT INTO todo (content, done, user_id) VALUES
('Buy milk', false, 1),
('Update documentation', false, 1),
('Fix roof', true, 1),
('Play games', true, 2),
('Rest', true, 2),
('Relax', true, 2),
('High five myself', false, 2);
