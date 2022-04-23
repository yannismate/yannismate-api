package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricMessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "twitchbot_messages_received",
		Help: "The total number of received messages",
	})
	metricChannelsJoined = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "twitchbot_channels_joined",
		Help: "Number of currently joined channels",
	})
	metricRankCommandsExecuted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "twitchbot_rank_commands_executed",
		Help: "Total number of rank commands executed",
	})
	metricRankCommandsCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "twitchbot_rank_commands_cache_hits",
		Help: "Total number of rank command cache hits",
	})
)
