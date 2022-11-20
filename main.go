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
	"net/http"
	"strconv"
	"time"
)

func history(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var latest map[string]string
	err := json.NewDecoder(req.Body).Decode(&latest)
	if err != nil {
		log.Error().Err(err).Msg("reading request failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	latestFromDevice := map[string]float64{}
	for deviceId, ts := range latest {
		t, err := strconv.ParseFloat(ts, 64)
		if err != nil {
			log.Error().Str("floatStr", ts).Err(err).Msg("invalid float in the request")
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		latestFromDevice[deviceId] = t
	}

	recs, err := storage.ReadEntries(db, latestFromDevice)
	if err != nil {
		log.Error().Err(err).Msg("reading records from DB failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	responseJson, err := json.Marshal(recs)
	if err != nil {
		log.Error().Err(err).Msg("marshalling json failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseJson)
	if err != nil {
		log.Error().Err(err).Msg("writing response failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func store(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var records []record.V1
	err := json.NewDecoder(req.Body).Decode(&records)
	if err != nil {
		log.Error().Err(err).Msg("reading request failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO remove - this is just a fake record
	records = append(records, record.V1{
		CmdLine:   "__test",
		Time:      fmt.Sprintf("%.4f", float64(time.Now().Unix())),
		DeviceID:  "test",
		SessionID: "test",
		RecordID:  "test1",
	})

	err = storage.StoreRecords(db, records)
	if err != nil {
		log.Error().Err(err).Msg("reading latest entry from the DB failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func latest(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var devices []string
	err := json.NewDecoder(req.Body).Decode(&devices)
	if err != nil {
		log.Error().Err(err).Msg("reading request failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	devicesSet := map[string]struct{}{}
	for _, device := range devices {
		devicesSet[device] = struct{}{}
	}
	lst, err := storage.LatestEntryPerDeviceId(db, devicesSet)
	if err != nil {
		log.Error().Err(err).Msg("reading latest entry from the DB failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseJson, err := json.Marshal(lst)
	if err != nil {
		log.Error().Err(err).Msg("marshalling json failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseJson)
	if err != nil {
		log.Error().Err(err).Msg("writing response failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
