CREATE TABLE users (
    id VARCHAR(27) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(31) NOT NULL,
    cell VARCHAR(31),
    picture JSONB NOT NULL,
    registration TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE INDEX index_users_on_registration ON users(registration);