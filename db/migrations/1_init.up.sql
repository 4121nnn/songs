CREATE TABLE IF NOT EXISTS songs (
   id UUID PRIMARY KEY,
   group_name VARCHAR(255) NOT NULL,
   song_name VARCHAR(255) NOT NULL,
   text TEXT,
   release_date VARCHAR(50),
   link VARCHAR(255)
);
