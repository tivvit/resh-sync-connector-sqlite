package main

import (
	"flag"
	"github.com/rs/zerolog/log"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/config"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/handlers"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/storage"
	"net/http"
)

func main() {
	configPath := flag.String("configPath", "conf.yml", "config path absolute or relative to binary")

	conf := config.New(*configPath)
	log.Info().
		Str("config file", *configPath).
		Interface("config", conf).
		Msg("configuration loaded")

	db, err := storage.ConnectDb(conf.SqlitePath)
	if err != nil {
		log.Fatal().Err(err).Str("path", conf.SqlitePath).Msg("connecting to DB failed")
	}

	http.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		handlers.Store(db, w, r)
	})
	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		handlers.History(db, w, r)
	})
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		handlers.Latest(db, w, r)
	})
	err = http.ListenAndServe(conf.Address, nil)
	if err != nil {
		panic(err)
	}
	err = db.Close()
	if err != nil {
		log.Error().Err(err).Msg("closing DB failed")
	}
}
