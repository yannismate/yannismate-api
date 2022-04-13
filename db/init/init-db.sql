CREATE DATABASE yannismate_api;

CREATE USER flyway PASSWORD 'flyway';
GRANT ALL PRIVILEGES ON DATABASE yannismate_api TO flyway;

CREATE USER twitch_bot PASSWORD 'twitch_bot';
GRANT ALL PRIVILEGES ON DATABASE yannismate_api TO twitch_bot;