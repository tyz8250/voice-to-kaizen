INSERT INTO users (name,email, password_hash, role)
VALUES (
  'Admin User',
  'admin@example.com',
  '$2a$10$wLSql2d68mAj6yvjkHaake5JnWBpIYHvCAkVBrowCQZat9ZRx8IXO',
  'admin'
)
ON CONFLICT (email) DO NOTHING;