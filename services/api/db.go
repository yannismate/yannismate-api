package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ApiDb struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

type ApiUser struct {
	UserId       int
	ApiKey       string
	RateLimit300 int
}

func NewApiDb(uri string) (*ApiDb, error) {
	ctx := context.Background()

	dbPool, err := pgxpool.Connect(ctx, uri)
	if err != nil {
		return nil, err
	}

	return &ApiDb{ctx: ctx, pool: dbPool}, nil
}

func (db *ApiDb) GetApiUserByKey(apiKey string) (*ApiUser, error) {

	var userId int32
	var ratelimit300 int32

	err := db.pool.QueryRow(db.ctx, "select user_id, api_key, ratelimit_300 from users_api where api_key=$1", apiKey).Scan(&userId, &apiKey, &ratelimit300)
	if err != nil {
		return nil, err
	}

	return &ApiUser{
		UserId:       int(userId),
		ApiKey:       apiKey,
		RateLimit300: int(ratelimit300),
	}, nil
}
