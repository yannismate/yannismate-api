package main

type Configuration struct {
	TwitchUsername       string `env:"TWITCH_USER"`
	TwitchToken          string `env:"TWITCH_TOKEN"`
	DbUri                string `json:"dbUri"`
	CacheUrl             string `json:"cacheUrl"`
	TrackerNetServiceUrl string `json:"trackerNetServiceUrl"`
}
