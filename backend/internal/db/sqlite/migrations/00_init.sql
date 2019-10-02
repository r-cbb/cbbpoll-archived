CREATE TABLE team
(
  id         INTEGER PRIMARY KEY,
  full_name  VARCHAR(64),
  short_name VARCHAR(32),
  nickname   VARCHAR(32),
  conference VARCHAR(32)
);

CREATE TABLE user
(
  nickname VARCHAR(32),
  is_admin BOOLEAN,
  is_voter BOOLEAN
);

CREATE TABLE poll
(
  season     INTEGER,
  week       INTEGER,
  week_name  VARCHAR(64),
  open_time  DATETIME,
  close_time DATETIME
);

CREATE TABLE ballot
(
  id           INTEGER PRIMARY KEY,
  updated_time DATETIME,
  user         VARCHAR(32),
  is_official  BOOLEAN
);

CREATE TABLE vote
(
  team_id INTEGER,
  rank    INTEGER,
  reason  VARCHAR(150)
);