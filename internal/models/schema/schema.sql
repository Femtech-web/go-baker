-- script to create user table 

DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  hashed_password CHAR(60) NOT NULL,
  isActive TINYINT(1) NOT NULL DEFAULT 0,
  created DATETIME NOT NULL,
  features JSON NULL
);

ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);

-- script to create session table 

DROP TABLE IF EXISTS sessions;
CREATE TABLE sessions (
  token CHAR(43) PRIMARY KEY,
  data BLOB NOT NULL,
  expiry TIMESTAMP(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- script to create features table 

DROP TABLE IF EXISTS features;
CREATE TABLE features (
  id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
  loaves  INTEGER NOT NULL,
  feature1  INTEGER NOT NULL,
  feature2  INTEGER NOT NULL,
  feature3  INTEGER NOT NULL,
  date DATETIME NOT NULL
);
-- Add an index on the date column.
CREATE INDEX idx_snippets_created ON features(date);