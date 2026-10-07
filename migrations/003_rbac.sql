CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh terhadap seluruh data komik dan sistem'),
    ('staff', 'Boleh mengelola seluruh data komik, tetapi tidak boleh mengubah role user'),
    ('user',  'Hanya boleh mengelola dan melihat komik milik sendiri')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
    ('comic:list',       'Melihat daftar seluruh komik'),
    ('comic:read:any',   'Melihat detail komik milik siapa pun'),
    ('comic:create',     'Menambahkan komik baru'),
    ('comic:update:any', 'Mengubah komik milik siapa pun'),
    ('comic:delete',     'Menghapus komik milik siapa pun'),
    ('role:assign',     'Mengubah role milik user lain')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'comic:list'),
    ('admin', 'comic:read:any'),
    ('admin', 'comic:create'),
    ('admin', 'comic:update:any'),
    ('admin', 'comic:delete'),
    ('admin', 'role:assign'),
    ('staff', 'comic:list'),
    ('staff', 'comic:read:any'),
    ('staff', 'comic:create'),
    ('staff', 'comic:update:any'),
    ('staff', 'comic:delete')
ON CONFLICT DO NOTHING;

UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);