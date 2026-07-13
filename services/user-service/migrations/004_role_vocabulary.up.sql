-- Fiberlane pivot: role vocabulary buyer/supplier -> startup/manufacturer.
-- (admin is unchanged.) The inline column CHECK from 003 is named users_role_check.
UPDATE users SET role = 'startup' WHERE role = 'buyer';
UPDATE users SET role = 'manufacturer' WHERE role = 'supplier';

ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('startup','manufacturer','admin'));
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'startup';
