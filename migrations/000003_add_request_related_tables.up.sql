BEGIN;

CREATE TABLE comments (
  id SERIAL PRIMARY KEY,
  request_id INTEGER NOT NULL REFERENCES kaizen_requests(id),
  author_id INTEGER NOT NULL REFERENCES users(id),
  body TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE status_histories (
  id SERIAL PRIMARY KEY,
  request_id INTEGER NOT NULL REFERENCES kaizen_requests(id),
  from_status TEXT NOT NULL,
  to_status TEXT NOT NULL,
  changed_by INTEGER NOT NULL REFERENCES users(id),
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE decision_logs (
  id SERIAL PRIMARY KEY,
  request_id INTEGER NOT NULL REFERENCES kaizen_requests(id),
  decided_by INTEGER NOT NULL REFERENCES users(id),
  decision TEXT NOT NULL,
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
