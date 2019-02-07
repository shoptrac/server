package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
	"gopkg.in/yaml.v2"
)

type LongPoll struct {
	ch (chan int)
}

type Endpoints struct {
	Rtr    *mux.Router
	lp     *LongPoll
	db     *mongo.Database
	JWTKey []byte
	JWTMW  *jwtmiddleware.JWTMiddleware
}

type Config struct {
	MongoURL  string `yaml:"mongo_url"`
	JWTSecret string `yaml:"jwt_secret"`
}

func NewEndPoints(r *mux.Router) *Endpoints {
	var c Config
	c.readConfig()

	database := c.connectMongo()
	secret, jwtmd := c.initJWT()

	ep := &Endpoints{
		Rtr:    r,
		lp:     nil,
		db:     database,
		JWTKey: secret,
		JWTMW:  jwtmd,
	}

	return ep
}

func (e *Endpoints) RegisterEndpoints() {

	// Dashboard Endpoints
	e.Rtr.Handle("/dashboard/control", e.JWTMW.Handler(http.HandlerFunc(e.sendControl))).Methods("GET")
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

func (c *Config) initJWT() ([]byte, *jwtmiddleware.JWTMiddleware) {
	jwtSecret := []byte(c.JWTSecret)

	jwtmw := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	return jwtSecret, jwtmw
}
