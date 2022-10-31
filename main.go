package main

import (
	"encoding/json"
	"fmt"
	"github.com/curusarn/resh/pkg/records"
	"math/rand"
	"net/http"
)

func history(w http.ResponseWriter, req *http.Request) {
	record := records.BaseRecord{
		CmdLine: fmt.Sprint("FAKE_TEST_", rand.Intn(100)),
		Host:    "__TEST__",
	}
	responseJson, err := json.Marshal(record)
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

func store(w http.ResponseWriter, req *http.Request) {
}

func latest(w http.ResponseWriter, req *http.Request) {
}

func main() {
	http.HandleFunc("/store", store)
	http.HandleFunc("/history", history)
	http.HandleFunc("/latest", latest)
	err := http.ListenAndServe(":1234", nil)
	if err != nil {
		panic(err)
	}
}
