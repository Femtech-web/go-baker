-- script to create user table 

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

CREATE TABLE sessions (
  token CHAR(43) PRIMARY KEY,
  data BLOB NOT NULL,
  expiry TIMESTAMP(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);