create table if not exists users_twitch (
    twitch_user_id varchar(20) not null,
    twitch_login varchar(30) not null,
    rl_platform varchar(20),
    rl_username varchar(255),
    rl_message_format varchar(255) not null default 'some format',
    primary key (twitch_user_id),
    unique (twitch_login)
);

create table if not exists users_api (
    user_id integer not null generated always as identity,
    api_key varchar(255) not null,
    ratelimit_300 integer not null,
    primary key (user_id),
    unique (api_key)
);