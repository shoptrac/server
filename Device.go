package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const ENTER = 0
const LEAVE = 1
const NUM_RETRIES = 3

// GET
func (e *Endpoints) openLongPoll(w http.ResponseWriter, r *http.Request) {

	// ch := make(chan int)
	// lp := &LongPoll{
	// 	ch: ch,
	// }

	// e.setLongPoll(lp)

	// controlCode := <-ch
	// json.NewEncoder(w).Encode(controlCode)
}

// Types used for Face++ requests
type Attributes struct {
	Sex map[string]string `json:"gender"`
	Age map[string]int    `json:"age"`
}

type Face struct {
	Atbs Attributes `json:"attributes"`
}

// POST
func (e *Endpoints) postEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := struct {
		DeviceID  string `json:"device_id"`
		Timestamp int64  `json:"timestamp"`
		Action    int    `json:"action"`
	}{}

	ret := struct {
		Success bool
	}{}
	ret.Success = true

	err := decoder.Decode(&params)

	if err != nil {
		ret.Success = false
		json.NewEncoder(w).Encode(ret)
		return
	}

	events := e.db.Collection("events")

	_, err = events.InsertOne(context.Background(),
		bson.M{"device_id": params.DeviceID, "action": params.Action, "timestamp": time.Unix(params.Timestamp/1000, 0)})

	if e.lp != nil {
		e.lp.ch <- "2" //
	} else {
		fmt.Println("No LP open")
	}

	json.NewEncoder(w).Encode(ret)
}

// POST
func (e *Endpoints) postImage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := struct {
		DeviceID  string `json:"device_id"`
		Timestamp int64  `json:"timestamp"`
		Image     string `json:"image"`
	}{}

	ret := struct {
		Success bool
	}{}
	ret.Success = true

	err := decoder.Decode(&params)
	profiles := e.db.Collection("profiles")

	attempt := 1

	// Attempt NUM_RETRIES times to send out the requests, sometimes there's random issues
	for attempt <= NUM_RETRIES {
		res, _ := http.PostForm(e.Fau, url.Values{"api_key": {e.Fak}, "api_secret": {e.Fas}, "return_attributes": {"gender,age"}, "image_base64": {params.Image}})

		defer res.Body.Close()

		body := struct {
			Faces []Face `json:"faces"`
			Error string `json:"error_message"`
		}{}

		decoder = json.NewDecoder(res.Body)
		err = decoder.Decode(&body)

		if err != nil || len(body.Error) > 0 {
			attempt += 1

			if attempt == 4 {
				ret.Success = false
				log.Fatal(body.Error)
				break
			}
			continue
		}

		if len(body.Faces) == 0 {
			ret.Success = false
			break
		}

		age := body.Faces[0].Atbs.Age["value"]
		sex := body.Faces[0].Atbs.Sex["value"]

		_, err = profiles.InsertOne(context.Background(),
			bson.M{"device_id": params.DeviceID, "timestamp": time.Unix(params.Timestamp/1000, 0), "age": age, "sex": sex})

		break
	}

	if e.lp != nil {
		e.lp.ch <- "1" //
	} else {
		fmt.Println("No LP open")
	}

	json.NewEncoder(w).Encode(ret)
}
