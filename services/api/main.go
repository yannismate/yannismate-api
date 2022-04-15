package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/tkanos/gonfig"
	"github.com/yannismate/yannismate-api/libs/cache"
	"github.com/yannismate/yannismate-api/libs/httplog"
	"github.com/yannismate/yannismate-api/libs/ratelimit"
	"net/http"
	"strconv"
	"time"
)

var configuration = Configuration{}
var ratelimiter ratelimit.SharedRateLimiter
var apiDb *ApiDb

func main() {
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		log.WithField("event", "load_config").Fatal(err)
		return
	}

	redisCache := cache.NewCache(configuration.CacheUrl)
	ratelimiter = ratelimit.NewSharedRateLimiter(&redisCache)

	apiDb, err = NewApiDb(configuration.DbUri)
	if err != nil {
		log.WithField("event", "connect_db").Fatal(err)
		return
	}

	http.Handle("/rank", httplog.WithLogging(withRateLimit(rankHandler())))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.WithField("event", "start_server").Fatal(err)
	}
}

func withRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("api_key")
		}
		if apiKey == "" {
			w.WriteHeader(403)
			return
		}

		limitRemaining, err := ratelimiter.AllowIfTracked("apikey:" + apiKey)
		if err != nil {

			apiUser, err := apiDb.GetApiUserByKey(apiKey)
			if err != nil {
				w.WriteHeader(403)
				return
			}

			limitRemaining, err = ratelimiter.AllowNew("apikey:"+apiKey, apiUser.RateLimit300, time.Second*300)
			if err != nil {
				log.WithField("event", "ratelimiter_allow_new").Error(err)
				w.WriteHeader(500)
				return
			}
			w.Header().Set("RateLimit-Remaining", strconv.Itoa(limitRemaining))

		} else {
			if limitRemaining < 0 {
				w.WriteHeader(429)
				return
			}
			w.Header().Set("RateLimit-Remaining", strconv.Itoa(limitRemaining))
		}
		next.ServeHTTP(w, r)
	})
}

func rankHandler() http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("hello"))
	}
	return http.HandlerFunc(fn)
}
