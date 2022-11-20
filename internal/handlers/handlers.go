package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/curusarn/resh/record"
	"github.com/rs/zerolog/log"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/storage"
	"net/http"
	"strconv"
)

func History(db *sql.DB, w http.ResponseWriter, req *http.Request) {
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

func Store(db *sql.DB, w http.ResponseWriter, req *http.Request) {
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

	err = storage.StoreRecords(db, records)
	if err != nil {
		log.Error().Err(err).Msg("reading latest entry from the DB failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Latest(db *sql.DB, w http.ResponseWriter, req *http.Request) {
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
	// add requested unknown deviceIds
	for deviceId, _ := range devicesSet {
		if _, ok := lst[deviceId]; !ok {
			lst[deviceId] = "0.0"
		}
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
