package main

import (
	"context"
	"encoding/json"	
	"log"
	"net/http"
	"time"

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
	
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	clientOptions := options.Client().ApplyURI(mongoAddress)
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()

	router.HandleFunc(observationEndpoint, CreateObservation).Methods("POST")

	http.ListenAndServe(":8090", router)
}