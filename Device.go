package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
    "fmt"

	"go.mongodb.org/mongo-driver/bson"
)

const ENTER = 0
const LEAVE = 1

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

// Types used for Face++ requests
type Attributes struct {
    Sex string `json:"gender"`
    Age int `json:"age"`
}

type Face struct {
   Atbs Attributes `json:"attributes"`
}


// POST
func (e *Endpoints) postEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := struct {
		DeviceID  string `json:"device_id"`
		Action    int    `json:"action"`
		Timestamp int64  `json:"timestamp"`
        Image     []byte `json:"image"`
	}{}

	ret := struct {
		Success bool
	}{}

	ret.Success = true

	err := decoder.Decode(&params)

	events := e.db.Collection("events")

	_, err = events.InsertOne(context.Background(),
		bson.M{"device_id": params.DeviceID, "action": params.Action, "timestamp": time.Unix(params.Timestamp/1000, 0)})

    if params.Action == ENTER {
	    json.NewEncoder(w).Encode(ret)
        return
    }

    url := e.Fau
    url += fmt.Sprintf("?api_key=%s", e.Fak)
    url += fmt.Sprintf("&api_secret=%s", e.Fas)
    url += "&return_attributes=gender,age"

    // This needs to be changed so that the POST posts a multipart form data request with the image byte data
    // API can be viewed at https://console.faceplusplus.com/documents/5679127
    res, _ := http.Post(url)
    defer res.Body.Close()

	body := struct {
		Faces   []Face `json:"faces"`
        Error   string `json:"error_message"`
	}{}

    decoder = json.NewDecoder(res.Body)
    err = decoder.Decode(&body)

    age = body.Faces[0].Atbs.Age
    sex = body.Faces[0].Atbs.Sex

	profiles := e.db.Collection("profiles")

	_, err = profiles.InsertOne(context.Background(),
		bson.M{"device_id": params.DeviceID, "action": params.Action, "timestamp": time.Unix(params.Timestamp/1000, 0),
        "age": age, "sex": sex})

	json.NewEncoder(w).Encode(ret)
}
