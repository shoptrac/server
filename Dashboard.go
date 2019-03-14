package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

type DateID struct {
	Year  int32 `bson:"year"`
	Month int32 `bson:"month"`
	Day   int32 `bson:"day"`
}

type HourID struct {
	Hour int32 `bson:"hour"`
}

type SexID struct {
	Sex string `bson:"sex"`
}

type DateSexID struct {
	Year  int32  `bson:"year"`
	Month int32  `bson:"month"`
	Day   int32  `bson:"day"`
	Sex   string `bson:"sex"`
}

type GAVPDResponse struct {
	ID    DateID `bson:"_id"`
	Count int    `bson:"count"`
}

type GPHResponse struct {
	ID    HourID `bson:"_id"`
	Count int    `bson:"count"`
}

type GSDResponse struct {
	ID    SexID `bson:"_id"`
	Count int   `bson:"count"`
}

type GADResponse struct {
	DeviceID string `bson:"device_id"`
	// Timestamp type `bson:"timestamp"`
	Age int    `bson:"age"`
	Sex string `bson:"sex"`
}

type GTHAgeResponse struct {
	ID     DateID `bson:"_id"`
	Count  int    `bson:"count"`
	AgeSum int    `bson:"age"`
}

type GTHSexResponse struct {
	ID    DateSexID `bson:"_id"`
	Count int       `bson:"count"`
}

// GET
func (e *Endpoints) openLongPollDashboard(w http.ResponseWriter, r *http.Request) {
	// check device id?
	ch := make(chan string)
	lp := &LongPoll{
		ch: ch,
	}

	e.setLongPoll(lp)
	controlCode := <-ch
	json.NewEncoder(w).Encode(controlCode)
}

// GET
func (e *Endpoints) sendControl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if e.lp != nil {
		e.lp.ch <- vars["signal"]
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
	ret.Data = (total / 2) / counter

	json.NewEncoder(w).Encode(ret)
}

// GET
func (e *Endpoints) getPeakHours(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Get average # of people in the store broken up by hours of the day (9-5)
	// Number of people in the store NOT how many people entered at that time

	// Hour will be in UTC so need to -5 from it.
	deviceId := "1111aaaa"
	eeMap := make([]map[string]int, 23)

	// Entering
	pl := bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}, {"action", 0}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"hour", bson.D{{"$hour", "$timestamp"}}}}}, {"count", bson.D{{"$sum", 1}}}}}}, bson.D{{"$sort", bson.D{{"_id.hour", 1}}}}}
	cur, err := e.db.Collection("events").Aggregate(context.Background(), pl)

	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GPHResponse{}

		err = cur.Decode(&elem)
		if err != nil {
			fmt.Println(err)
		}

		eeMap[elem.ID.Hour-5] = make(map[string]int)
		eeMap[elem.ID.Hour-5]["enter"] = elem.Count
	}

	// Exiting
	pl = bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}, {"action", 1}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"hour", bson.D{{"$hour", "$timestamp"}}}}}, {"count", bson.D{{"$sum", 1}}}}}}, bson.D{{"$sort", bson.D{{"_id.hour", 1}}}}}
	cur, err = e.db.Collection("events").Aggregate(context.Background(), pl)

	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GPHResponse{}

		err = cur.Decode(&elem)
		if err != nil {
			// error
		}

		if eeMap[elem.ID.Hour-5] == nil {
			
			eeMap[elem.ID.Hour-5] = make(map[string]int)
		}
	
		eeMap[elem.ID.Hour-5]["exit"] = elem.Count
		
	}

	carryOver := 0
	trafficMap := make(map[int]int)
	for i := range eeMap {
		if eeMap[i] != nil {
			trafficMap[i] = eeMap[i]["enter"] + carryOver
			carryOver = carryOver + (eeMap[i]["enter"] - eeMap[i]["exit"])
		}
	}

	ret := struct {
		Data    map[int]int
		Success bool
	}{}
	ret.Data = trafficMap
	ret.Success = true

	json.NewEncoder(w).Encode(ret)
}

// GET
func (e *Endpoints) getTrafficHistory(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Get a table with history (by day) of how many people, what kind of people came into the store. Past ~10 days?

	deviceId := "1111aaaa"

	retData := make(map[string]map[string]int)

	pl := bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"year", bson.D{{"$year", "$timestamp"}}}, {"month", bson.D{{"$month", "$timestamp"}}}, {"day", bson.D{{"$dayOfMonth", "$timestamp"}}}}}, {"count", bson.D{{"$sum", 1}}}, {"age", bson.D{{"$sum", "$age"}}}}}}}

	cur, err := e.db.Collection("profiles").Aggregate(context.Background(), pl)

	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GTHAgeResponse{}
		retDataPnt := make(map[string]int)

		err = cur.Decode(&elem)

		if err != nil {
			// error
		}

		retDataPnt["count"] = elem.Count
		retDataPnt["avgAge"] = elem.AgeSum / elem.Count

		retData[fmt.Sprintf("%d-%d-%d", elem.ID.Day, elem.ID.Month, elem.ID.Year)] = retDataPnt
	}

	pl = bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"year", bson.D{{"$year", "$timestamp"}}}, {"month", bson.D{{"$month", "$timestamp"}}}, {"day", bson.D{{"$dayOfMonth", "$timestamp"}}}, {"sex", "$sex"}}}, {"count", bson.D{{"$sum", 1}}}}}}}
	cur, err = e.db.Collection("profiles").Aggregate(context.Background(), pl)

	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GTHSexResponse{}
		err = cur.Decode(&elem)

		if err != nil {
			// error
		}

		if elem.ID.Sex == "Female" {
			retData[fmt.Sprintf("%d-%d-%d", elem.ID.Day, elem.ID.Month, elem.ID.Year)]["females"] = elem.Count
		} else if elem.ID.Sex == "Male" {
			retData[fmt.Sprintf("%d-%d-%d", elem.ID.Day, elem.ID.Month, elem.ID.Year)]["males"] = elem.Count
		}
	}

	ret := struct {
		Data    map[string]map[string]int
		Success bool
	}{}
	ret.Data = retData
	ret.Success = true

	json.NewEncoder(w).Encode(ret)
}

// GET
func (e *Endpoints) getSexDist(w http.ResponseWriter, r *http.Request) {
	deviceId := "1111aaaa"
	retData := make(map[string]int)

	pl := bson.A{bson.D{{"$match", bson.D{{"device_id", deviceId}}}}, bson.D{{"$group", bson.D{{"_id", bson.D{{"sex", "$sex"}}}, {"count", bson.D{{"$sum", 1}}}}}}}
	cur, err := e.db.Collection("profiles").Aggregate(context.Background(), pl)

	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GSDResponse{}

		err = cur.Decode(&elem)
		if err != nil {
			fmt.Println(err)
		}

		retData[elem.ID.Sex] = elem.Count
	}

	ret := struct {
		Success bool
		Data    map[string]int
	}{}
	ret.Success = true
	ret.Data = retData

	json.NewEncoder(w).Encode(ret)
}

func (e *Endpoints) getAgeDist(w http.ResponseWriter, r *http.Request) {
	deviceId := "1111aaaa"

	filter := bson.D{{"device_id", deviceId}}

	retData := make(map[int]int)

	cur, err := e.db.Collection("profiles").Find(context.Background(), filter)

	if err != nil {
		fmt.Println(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		elem := GADResponse{}

		err = cur.Decode(&elem)
		if err != nil {
			fmt.Println(err)
		}

		retData[elem.Age/10] += 1
	}

	ret := struct {
		Success bool
		Data    map[int]int
	}{}
	ret.Success = true
	ret.Data = retData

	json.NewEncoder(w).Encode(ret)
}

// DEMO Current number of people in the store !
// DEMO Get Last profile
