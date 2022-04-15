CREATE ROLE apps NOLOGIN;
CREATE DATABASE yannismate_api WITH OWNER apps;

CREATE USER flyway PASSWORD 'flyway';
GRANT ALL PRIVILEGES ON DATABASE yannismate_api TO flyway;

CREATE USER twitch_bot PASSWORD 'twitch_bot';
GRANT apps TO twitch_bot;

CREATE USER api PASSWORD 'api';
GRANT apps TO api;