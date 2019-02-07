package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

// GET
func (e *Endpoints) sendControl(w http.ResponseWriter, r *http.Request) {
	if e.lp != nil {
		e.lp.ch <- 1
		json.NewEncoder(w).Encode("Successfully closed")
	} else {
		json.NewEncoder(w).Encode("No LP open")
	}
}

// POST
func (e *Endpoints) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := struct {
		Email    string `json:email`
		Password string `json:password`
	}{}

	mgoResult := struct {
		Id       string `json:_id`
		Email    string `json:email`
		Password string `json:password`
	}{}

	err := decoder.Decode(&params)

	if err != nil {
		log.Fatal(err)
	}

	collection := e.db.Collection("users")

	filter := bson.M{"email": params.Email, "password": params.Password}
	err = collection.FindOne(context.Background(), filter).Decode(&mgoResult)
}
