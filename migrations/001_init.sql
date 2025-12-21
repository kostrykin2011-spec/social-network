CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY NOT NULL,
	password VARCHAR(255) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS profiles (
        id UUID PRIMARY KEY NOT NULL,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	last_name VARCHAR(100) NOT NULL,
	first_name VARCHAR(100) NOT NULL,
	birth_date DATE NOT NULL,
	gender VARCHAR(20) NOT NULL,
	biography TEXT,
	city VARCHAR(100),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS friendships (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        friend_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        status VARCHAR(20) DEFAULT 'pending',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(user_id, friend_id)
);
CREATE TABLE IF NOT EXISTS posts (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        is_public BOOLEAN DEFAULT false,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_profiles_fullname_btree ON profiles (
	last_name varchar_pattern_ops, 
	first_name varchar_pattern_ops
);
