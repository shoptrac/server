package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
)

type DateID struct {
	Year  int32 `bson:"year"`
	Month int32 `bson:"month"`
	Day   int32 `bson:"day"`
}

type GAVPDResponse struct {
	ID    DateID `bson:"_id"`
	Count int    `bson:"count"`
}

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
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	var mgoResult struct {
		Email       string `bson:"email"`
		DisplayName string `bson:"display_name"`
		Password    string `bson:"password"`
	}

	ret := struct {
		DisplayName string
		Token       string
		Success     bool
	}{}

	err := decoder.Decode(&params)

	if err != nil {
		log.Fatal(err)
	}

	collection := e.db.Collection("users")
	filter := bson.D{{"email", params.Email}, {"password", params.Password}}
	err = collection.FindOne(context.Background(), filter).Decode(&mgoResult)

	if mgoResult.Email == params.Email {
		// Login Success
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["email"] = mgoResult.Email // Should be _id but driver won't return in
		// claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Too lazy to make it expire

		tokenString, _ := token.SignedString(e.JWTKey)
		ret.DisplayName = mgoResult.DisplayName
		ret.Token = tokenString
		ret.Success = true

		json.NewEncoder(w).Encode(ret)
	} else {
		// Login Failure
		ret.Success = false

		json.NewEncoder(w).Encode(ret)
	}
}

// GET
func (e *Endpoints) getAverageDuration(w http.ResponseWriter, r *http.Request) {
	// TODO
	// get the average user visit duration (single value)
}

// GET
func (e *Endpoints) getAverageVisitsPD(w http.ResponseWriter, r *http.Request) {
	// get the average number of visits per day (single value)

	// vars := mux.Vars(r) / get the deviceID
	deviceId := "1111aaaa"

	pl := bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"year", bson.D{{"$year", "$timestamp"}}}, {"month", bson.D{{"$month", "$timestamp"}}}, {"day", bson.D{{"$dayOfMonth", "$timestamp"}}}}}, {"count", bson.D{{"$sum", 1}}}}}}}

	cur, err := e.db.Collection("events").Aggregate(context.Background(), pl)

	if err != nil {
		fmt.Println("Error")
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	counter := 0
	total := 0
	for cur.Next(context.Background()) {
		elem := GAVPDResponse{}

		err = cur.Decode(&elem)

		if err != nil {
			// error
		}

		total += elem.Count
		counter++ // increment counter for every day
	}

	ret := struct {
		Success bool
		Data    int
	}{}
	ret.Success = true
	ret.Data = total / counter

	json.NewEncoder(w).Encode(ret)
}

// GET
func (e *Endpoints) getPeakHours(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Get average # of people in the store broken up by hours of the day (9-5)
}

// GET
func (e *Endpoints) getTrafficHistory(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Get a table with history (by day) of how many people, what kind of people came into the store. Past ~10 days?
}

// DEMO Current number of people in the store
// DEMO Get Last
