CREATE TABLE kaizen_requests (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  category TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN (
    'captured',
    'owner_needed',
    'planned',
    'in_progress',
    'done',
    'rejected'
  )),
  impact INTEGER NOT NULL CHECK (impact BETWEEN 1 AND 5),
  urgency INTEGER NOT NULL CHECK (urgency BETWEEN 1 AND 5),
  effort INTEGER NOT NULL CHECK (effort BETWEEN 1 AND 5),
  priority_score INTEGER NOT NULL,
  requester_id INTEGER NOT NULL,
  owner_id INTEGER,
  next_action TEXT,
  due_date DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_requester
    FOREIGN KEY (requester_id)
    REFERENCES users(id),
  CONSTRAINT fk_owner
    FOREIGN KEY (owner_id)
    REFERENCES users(id)
);
