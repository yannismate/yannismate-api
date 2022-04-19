package main

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type BotDb struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

type BotUser struct {
	TwitchUserId      string
	TwitchLogin       string
	TwitchCommandName string
	RlPlatform        string
	RlUsername        string
	RlMessageFormat   string
}

func NewBotDb(uri string) (*BotDb, error) {
	ctx := context.Background()

	dbPool, err := pgxpool.Connect(ctx, uri)
	if err != nil {
		return nil, err
	}

	return &BotDb{ctx: ctx, pool: dbPool}, nil
}

func (db *BotDb) GetBotUserByTwitchUserId(twitchUserId string) (*BotUser, error) {
	row := db.pool.QueryRow(db.ctx, `select twitch_user_id, twitch_login, twitch_command_name, rl_platform, rl_username, 
		rl_message_format from users_twitch where twitch_user_id=$1;`, twitchUserId)
	return toBotUser(row)
}

func (db *BotDb) GetBotUserByTwitchLogin(twitchLogin string) (*BotUser, error) {
	row := db.pool.QueryRow(db.ctx, `select twitch_user_id, twitch_login, twitch_command_name, rl_platform, rl_username, rl_message_format 
		from users_twitch where twitch_login=$1;`, twitchLogin)
	return toBotUser(row)
}

func (db *BotDb) UpdateRlPlatformAndUsernameByTwitchLogin(twitchLogin string, rlPlatform string, rlUsername string) (bool, error) {
	cmdTag, err := db.pool.Exec(db.ctx, `update users_twitch set rl_platform=$1, rl_username=$2 where twitch_login=$3;`, rlPlatform, rlUsername, twitchLogin)
	return cmdTag.RowsAffected() > 0, err
}

func (db *BotDb) UpdateRlPlatformByTwitchLogin(twitchLogin string, rlPlatform string) (bool, error) {
	cmdTag, err := db.pool.Exec(db.ctx, `update users_twitch set rl_platform=$1 where twitch_login=$2;`, rlPlatform, twitchLogin)
	return cmdTag.RowsAffected() > 0, err
}

func (db *BotDb) UpdateRlUsernameByTwitchLogin(twitchLogin string, rlUsername string) (bool, error) {
	cmdTag, err := db.pool.Exec(db.ctx, `update users_twitch set rl_username=$1 where twitch_login=$2;`, rlUsername, twitchLogin)
	return cmdTag.RowsAffected() > 0, err
}

func (db *BotDb) UpdateRlMsgFormatByTwitchLogin(twitchLogin string, format string) (bool, error) {
	cmdTag, err := db.pool.Exec(db.ctx, `update users_twitch set rl_message_format=$1 where twitch_login=$2;`, format, twitchLogin)
	return cmdTag.RowsAffected() > 0, err
}

func (db *BotDb) UpdateTwitchCommandNameByTwitchLogin(twitchLogin string, cmd string) (bool, error) {
	cmdTag, err := db.pool.Exec(db.ctx, `update users_twitch set twitch_command_name=$1 where twitch_login=$2;`, cmd, twitchLogin)
	return cmdTag.RowsAffected() > 0, err
}

func (db *BotDb) InsertBotUser(user BotUser) error {
	_, err := db.pool.Exec(db.ctx, `insert into users_twitch (twitch_user_id, twitch_login, twitch_command_name, rl_platform, rl_username, rl_message_format)
		values ($1, $2, $3, $4, $5, $6);`, user.TwitchUserId, user.TwitchLogin, user.TwitchCommandName, user.RlPlatform, user.RlUsername, user.RlMessageFormat)
	return err
}

func (db *BotDb) DeleteBotUserByTwitchUserId(twitchUserId string) (bool, error) {
	cmdTag, err := db.pool.Exec(db.ctx, `delete from users_twitch where twitch_user_id=$1;`, twitchUserId)
	return cmdTag.RowsAffected() > 0, err
}

func (db *BotDb) GetUserNames(userNameGt string, pageSize int) ([]string, *string, error) {
	rows, err := db.pool.Query(db.ctx, `select twitch_login from users_twitch where twitch_login > $1 
		order by twitch_login asc limit $2;`, userNameGt, pageSize)
	loginNames := make([]string, 0)
	if err != nil {
		return loginNames, nil, err
	}

	for rows.Next() {
		var userLogin string
		err = rows.Scan(&userLogin)
		if err != nil {
			return loginNames, nil, err
		}
		loginNames = append(loginNames, userLogin)
	}

	if len(loginNames) > 0 {
		return loginNames, &loginNames[len(loginNames)-1], nil
	}

	return loginNames, nil, nil
}

func toBotUser(row pgx.Row) (*BotUser, error) {
	user := BotUser{}
	err := row.Scan(&user.TwitchUserId, &user.TwitchLogin, &user.TwitchCommandName, &user.RlPlatform, &user.RlUsername, &user.RlMessageFormat)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
