package main

type Configuration struct {
	TwitchUsername string          `env:"TWITCH_USER"`
	TwitchToken    string          `env:"TWITCH_TOKEN"`
	RateLimits     RateLimitConfig `json:"rateLimits"`
}

type RateLimitConfig struct {
	JoinsPer10    int `json:"joinsPer10"`
	MessagesPer30 int `json:"messagesPer30"`
}
