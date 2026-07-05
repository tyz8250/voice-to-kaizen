CREATE TABLE kaizen_requests (
    id SERIAL PRIMARY KEY,
    requester_id INTEGER NOT NULL,
    owner_id INTEGER,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'in_progress', 'closed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_requester 
        FOREIGN KEY (requester_id) 
        REFERENCES users(id),
    CONSTRAINT fk_owner 
        FOREIGN KEY (owner_id) 
        REFERENCES users(id)
);
