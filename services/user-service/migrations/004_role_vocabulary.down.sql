ALTER TABLE users ALTER COLUMN role SET DEFAULT 'buyer';
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('buyer','supplier','admin'));

UPDATE users SET role = 'supplier' WHERE role = 'manufacturer';
UPDATE users SET role = 'buyer' WHERE role = 'startup';
