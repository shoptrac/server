package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
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
	// TODO
	// get the average number of visits per day (single value)
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
