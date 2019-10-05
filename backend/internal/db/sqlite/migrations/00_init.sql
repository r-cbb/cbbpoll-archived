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
  nickname     VARCHAR(32),
  is_admin     BOOLEAN,
  is_voter     BOOLEAN,
  primary_team INTEGER,
  PRIMARY KEY (nickname),
  FOREIGN KEY (primary_team) REFERENCES team (id)
);

CREATE TABLE poll
(
  season        INTEGER,
  week          INTEGER,
  week_name     VARCHAR(64),
  open_time     DATETIME,
  close_time    DATETIME,
  last_modified DATETIME,
  reddit_url    TEXT,
  PRIMARY KEY (season, week)
);

CREATE TABLE ballot
(
  id           INTEGER PRIMARY KEY,
  poll_season  INTEGER,
  poll_week    INTEGER,
  updated_time DATETIME,
  user         VARCHAR(32),
  is_official  BOOLEAN,
  FOREIGN KEY (poll_season, poll_week) REFERENCES poll (season, week),
  FOREIGN KEY (user) REFERENCES user (nickname)
);

CREATE TABLE vote
(
  ballot_id INTEGER,
  team_id   INTEGER,
  rank      INTEGER,
  reason    VARCHAR(150),
  FOREIGN KEY (team_id) REFERENCES team (id)
);

CREATE TABLE result
(
  poll_season       INTEGER,
  poll_week         INTEGER,
  team_id           INTEGER,
  team_name         VARCHAR(64),
  team_slug         VARCHAR(32),
  rank              INTEGER,
  first_place_votes INTEGER,
  points            INTEGER,
  official          BOOLEAN,
  FOREIGN KEY (poll_season, poll_week) REFERENCES poll (season, week)
)