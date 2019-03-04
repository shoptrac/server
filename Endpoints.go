package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	e.Rtr.HandleFunc("/dashboard/login", e.loginUser).Methods("POST")
	e.Rtr.Handle("/dashboard/control", e.JWTMW.Handler(http.HandlerFunc(e.sendControl))).Methods("GET")
	e.Rtr.Handle("/dashboard/averageDuration", e.JWTMW.Handler(http.HandlerFunc(e.getAverageDuration))).Methods("GET")
	e.Rtr.Handle("/dashboard/averageVisitsPD", e.JWTMW.Handler(http.HandlerFunc(e.getAverageVisitsPD))).Methods("GET")
	e.Rtr.Handle("/dashboard/peakHours", e.JWTMW.Handler(http.HandlerFunc(e.getPeakHours))).Methods("GET")
	e.Rtr.Handle("/dashboard/history", e.JWTMW.Handler(http.HandlerFunc(e.getTrafficHistory))).Methods("GET")

	// Device Endpoints
	e.Rtr.HandleFunc("/device/olp", e.openLongPoll).Methods("GET")
	e.Rtr.HandleFunc("/device/event", e.postEvent).Methods("POST")
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
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(c.MongoURL))

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
