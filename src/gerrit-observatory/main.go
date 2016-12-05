package main

import (
	"gerrit-observatory/gerrit"
	"gerrit-observatory/observer"
	"gerrit-observatory/redis"
	"gerrit-observatory/http"
)

var (
	GerritPort  = 29418
	GerritUser  = "autodeploy"
	GerritHost  = "git.wandoulabs.com"
	PrivateKey  = "./.id_rsa"
	RedisHost   = "127.0.0.1"
	RedisPort   = 6379
	RedisDB     = 0
	PostTimeout = 60
)

// Config global config
type Config struct {
	GerritPort  int
	GerritUser  string
	GerritHost  string
	PrivateKey  string
	RedisHost   string
	RedisPort   int
	RedisDB     int
	PostTimeout int
}

func main() {
	config := &Config{
		GerritPort:  GerritPort,
		GerritUser:  GerritUser,
		GerritHost:  GerritHost,
		PrivateKey:  PrivateKey,
		RedisHost:   RedisHost,
		RedisPort:   RedisPort,
		RedisDB:     RedisDB,
		PostTimeout: PostTimeout,
	}

	redis.InitRedis(config.RedisHost, config.RedisPort, config.RedisDB)
	defer redis.DestroyRedis()

	eventStream, err := gerrit.NewEventStream(config.GerritPort, config.GerritUser, config.GerritHost, config.PrivateKey)
	if err != nil {
		panic(err)
	}
	observerContr := observer.NewObserverContr(eventStream.Channel, config.PostTimeout)
	subscribes, err := redis.GetSubscribes()
	if err != nil {
		panic(err)
	}
	for _, obs := range subscribes {
		err = observerContr.AddObserver(obs)
		if err != nil {
			panic(err)
		}
	}
	go observerContr.Start()
	go eventStream.Run()

	http.Router()
}
