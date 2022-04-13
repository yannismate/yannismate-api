create table if not exists users_twitch (
    twitch_user_id varchar(20) not null,
    twitch_login varchar(30) not null,
    rl_platform varchar(20),
    rl_username varchar(255),
    rl_message_format varchar(255) not null default 'some format',
    primary key (twitch_user_id),
    unique (twitch_login)
)