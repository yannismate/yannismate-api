package main

type Configuration struct {
	TrackerNet TnConfig
	Cache      CacheConfig
	ScraperUrl string
}

type TnConfig struct {
	BaseUrl string
}

type CacheConfig struct {
	RedisUrl   string
	TtlSeconds int
}
