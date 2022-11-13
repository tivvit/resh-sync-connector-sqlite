package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/curusarn/resh/record"
	"github.com/rs/zerolog/log"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/config"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/storage"
	"math/rand"
	"net/http"
	"time"
)

func history(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	// TODO read request
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var recs []record.V1
	recs = append(recs, record.V1{
		CmdLine: fmt.Sprint("FAKE_TEST_", rand.Intn(1000)),
		Device:  "__TEST__",
		Time:    fmt.Sprintf("%.4f", float64(time.Now().Unix())),
	})
	responseJson, err := json.Marshal(recs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseJson)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func store(db *sql.DB, w http.ResponseWriter, req *http.Request) {
}

func latest(db *sql.DB, w http.ResponseWriter, req *http.Request) {

}

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
		store(db, w, r)
	})
	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		history(db, w, r)
	})
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		latest(db, w, r)
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
