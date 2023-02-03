CREATE TABLE posts (
    id serial PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    npub VARCHAR(100) NOT NULL,
    relaylist TEXT NOT NULL,
    body TEXT NOT NULL
);
