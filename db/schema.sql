CREATE TABLE
    articles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL UNIQUE,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        title TEXT NOT NULL,
        textMD TEXT NOT NULL
    );

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL UNIQUE,
    registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_admin INTEGER DEFAULT 0 NOT NULL, -- boolean 0/1
);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    session_key_hash BLOB NOT NULL UNIQUE, -- hashed UUID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_use DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER DEFAULT 1 NOT NULL, -- boolean 0/1
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE invites (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        code TEXT NOT NULL, -- UUID string
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        claimed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        is_active INTEGER DEFAULT 1 NOT NULL, -- boolean 0/1
        claimed_by_user_id INTEGER DEFAULT 1 NOT NULL,
        FOREIGN KEY (claimed_by_user_id) REFERENCES users (id)
    );