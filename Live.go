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

// GET
func (e *Endpoints) getCurrentCount(w http.ResponseWriter, r *http.Request) {
	deviceId := "1111aaaa"

	now := time.Now().UTC()

	// Entering
	pl := bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}, {"action", 0}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"year", bson.D{{"$year", "$timestamp"}}}, {"month", bson.D{{"$month", "$timestamp"}}}, {"day", bson.D{{"$dayOfMonth", "$timestamp"}}}}}, {"count", bson.D{{"$sum", 1}}}}}}}

	cur, err := e.db.Collection("events").Aggregate(context.Background(), pl)
	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	counter := 0

	for cur.Next(context.Background()) {
		elem := GAVPDResponse{}
		err = cur.Decode(&elem)

		if err != nil {
			fmt.Println(err)
		}

		if elem.ID.Year == int32(now.Year()) && elem.ID.Month == int32(now.Month()) && elem.ID.Day == int32(now.Day()) {
			counter = elem.Count
		}

	}

	// leaving
	pl = bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}, {"action", 1}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"year", bson.D{{"$year", "$timestamp"}}}, {"month", bson.D{{"$month", "$timestamp"}}}, {"day", bson.D{{"$dayOfMonth", "$timestamp"}}}}}, {"count", bson.D{{"$sum", 1}}}}}}}

	cur, err = e.db.Collection("events").Aggregate(context.Background(), pl)
	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GAVPDResponse{}
		err = cur.Decode(&elem)

		if err != nil {
			fmt.Println(err)
		}

		if elem.ID.Year == int32(now.Year()) && elem.ID.Month == int32(now.Month()) && elem.ID.Day == int32(now.Day()) {
			counter -= elem.Count
		}

	}

	ret := struct {
		Data    int
		Success bool
	}{}
	ret.Success = true
	ret.Data = counter

	json.NewEncoder(w).Encode(ret)
}
