package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type lpData struct {
	TimeStamp time.Time `bson:"timestamp"`
	Age       int32     `bson:"age"`
	Sex       string    `bson:"sex"`
}

// GET
func (e *Endpoints) getLatestProfile(w http.ResponseWriter, r *http.Request) {
	deviceId := "1111aaaa"

	var mgoResult lpData

	ret := struct {
		Data    lpData
		Success bool
	}{}
	ret.Success = true

	collection := e.db.Collection("profiles")

	filter := bson.D{{"device_id", deviceId}}
	opts := &options.FindOneOptions{
		Sort: bson.D{{"$natural", -1}},
	}
	err := collection.FindOne(context.Background(), filter, opts).Decode(&mgoResult)

	if err != nil {
		fmt.Println(err)
	}

	ret.Data = mgoResult
	json.NewEncoder(w).Encode(ret)
}
