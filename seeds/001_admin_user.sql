INSERT INTO users (email, password_hash, role)
VALUES (
  'admin@example.com',
  '$2a$10$wLSql2d68mAj6yvjkHaake5JnWBpIYHvCAkVBrowCQZat9ZRx8IXO',
  'admin'
)
ON CONFLICT (email) DO NOTHING;