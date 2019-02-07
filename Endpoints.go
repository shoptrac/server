package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
	"gopkg.in/yaml.v2"
)

type LongPoll struct {
	ch (chan int)
}

type Endpoints struct {
	Rtr *mux.Router
	lp  *LongPoll
	db  *mongo.Database
}

type Config struct {
	MongoURL string `yaml:"mongo_url"`
}

func NewEndPoints(r *mux.Router) *Endpoints {
	var c Config
	c.readConfig()

	database := c.connectMongo()

	ep := &Endpoints{
		Rtr: r,
		lp:  nil,
		db:  database,
	}

	return ep
}

func (e *Endpoints) RegisterEndpoints() {

	// Dashboard Endpoints
	e.Rtr.HandleFunc("/dashboard/control", e.sendControl).Methods("GET")
	e.Rtr.HandleFunc("/dashboard/login", e.loginUser).Methods("POST")

	// Device Endpoints
	e.Rtr.HandleFunc("/device/olp", e.openLongPoll).Methods("GET")
}

func (e *Endpoints) setLongPoll(lp *LongPoll) {
	e.lp = lp
}

// Config stuff
func (c *Config) readConfig() *Config {
	file, err := ioutil.ReadFile("main.yaml")

	if err != nil {
		log.Printf("Error reading from main.yaml: %v", err)
	}

	err = yaml.Unmarshal(file, c)

	if err != nil {
		log.Fatalf("Error unmarshalling yaml: %v", err)
	}
	return c
}

func (c *Config) connectMongo() *mongo.Database {
	client, err := mongo.NewClient(c.MongoURL)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}

	database := client.Database("ShopTrac") /*.Collection("users")*/

	return database
}
