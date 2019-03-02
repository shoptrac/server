package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
)

// GET
func (e *Endpoints) openLongPoll(w http.ResponseWriter, r *http.Request) {

	ch := make(chan int)
	lp := &LongPoll{
		ch: ch,
	}

	e.setLongPoll(lp)

	controlCode := <-ch
	json.NewEncoder(w).Encode(controlCode)
}

// POST
func (e *Endpoints) postEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := struct {
		DeviceID  string `json:"device_id"`
		Action    int    `json:"action"`
		Timestamp int64  `json:"timestamp"`
	}{}

	ret := struct {
		Success bool
	}{}
	ret.Success = true

	err := decoder.Decode(&params)

	if err != nil {
		log.Fatal(err)
	}

	collection := e.db.Collection("events")

	_, err = collection.InsertOne(context.Background(),
		bson.M{"device_id": params.DeviceID, "action": params.Action, "timestamp": time.Unix(params.Timestamp/1000, 0)})

	if err != nil {
		ret.Success = false
	}

	// ADD WEBSOCKETS FOR LIVE DEMO

	json.NewEncoder(w).Encode(ret)
}
