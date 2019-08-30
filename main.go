package main

import (
	"context"
	"encoding/json"	
	"log"
	"net/http"
	"time"
	"os"
    "os/signal"

	"github.com/gorilla/mux"
	
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

const (
	mongoAddress string = "mongodb://localhost:27017"
	observationEndpoint string = "/observation"
	defaultTimeout time.Duration =  10 * time.Second
)

type Metric struct {
	Description string `json:"description" bson:"description"` 
	Measurement int64  `json:"measurement" bson:"measurement"`
}

type Observation struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Label     string   `json:"label" bson:"label"`
	Timestamp int64    `json:"timestamp" bson:"timestamp"`
	Values    []Metric `json:"values" bson:"values"`
}

func CreateObservation(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	var observation Observation
	_ = json.NewDecoder(req.Body).Decode(&observation)
	collection := client.Database("AggregatorDB").Collection("observations")
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	result, _ := collection.InsertOne(ctx, observation)

	json.NewEncoder(resWriter).Encode(result)
}

func GetObservation(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	log.Println("Request for " + observationEndpoint + " received." + req)

	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var observation Observation

	collection := client.Database("AggregatorDB").Collection("observations")
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	err := collection.FindOne(ctx, Observation{ID: id}).Decode(&observation)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`))
	}
}

func main() {
	log.Println("Starting service.")
	
	var wait time.Duration
    flag.DurationVar(&wait, "graceful-timeout", time.Second * 15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()
	
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoAddress)
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()


	router.HandleFunc(observationEndpoint, CreateObservation).Methods("POST")
	router.HandleFunc(observationEndpoint, GetObservation).Methods("GET")

	srv := &http.Server{
		Addr:         "0.0.0.0:8090",        
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: router, 
	}
	
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

	srv.Shutdown(ctx)

	log.Println("Shutting down service.")

	os.Exit(0)	
}