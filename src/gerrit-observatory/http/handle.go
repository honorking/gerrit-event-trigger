package http

import (
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/gorilla/mux"

	"gerrit-observatory/redis"
)

func Router() {
	r := mux.NewRouter()

	r.HandleFunc("/observers", ObserversPostHandler).Methods("POST")
	r.HandleFunc("/observers", ObserversGetHandler).Methods("GET")
	r.HandleFunc("/observers/{observerId}", ObserverGetHandler).Methods("GET")
	r.HandleFunc("/observers/{observerId}", ObserverDeleteHandler).Methods("DELETE")

	http.ListenAndServe(":8080", r)
}

func ObserversPostHandler(w http.ResponseWriter, r *http.Request) {
	var req redis.SubscribeDetail

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	req.Save()
	return
}

func ObserversGetHandler(w http.ResponseWriter, r *http.Request) {
	subscribes, err := redis.GetSubscribes()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	b, _ := json.Marshal(subscribes)
	w.Write(b)
	return
}

func ObserverGetHandler(w http.ResponseWriter, r *http.Request) {
	observerId, err := strconv.Atoi(mux.Vars(r)["observerId"])
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	subscribe, err := redis.GetSubscribe(observerId)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	b, _ := json.Marshal(subscribe)
	w.Write(b)
	return
}

func ObserverDeleteHandler(w http.ResponseWriter, r *http.Request) {
	observerId, err := strconv.Atoi(mux.Vars(r)["observerId"])
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	subscribe, err := redis.DeleteSubscribe(observerId)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	b, _ := json.Marshal(subscribe)
	w.Write(b)
	return
}
