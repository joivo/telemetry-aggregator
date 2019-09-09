package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

const (
	dbAddress           string = "mongodb://db:27017"
	dbName              string = "aggregatordb"
	dbCollection        string = "observations"
	observationEndpoint string = "/observation"
	versionEndpoint     string = "/version"
	defaultTimeout             = 10 * time.Second
)

type Version struct {
	V string `json:"version"`
}

type Metric struct {
	Description string `json:"description" bson:"description"`
	Measurement float64  `json:"measurement" bson:"measurement"`
}

type Observation struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Label     string             `json:"label" bson:"label"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
	Values    []Metric           `json:"values" bson:"values"`
}

func GetVersion(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(resWriter).Encode(Version{"1.0.0"}); err != nil {
		log.Println(err.Error())
	}
}

func CreateObservation(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	var observation Observation
	_ = json.NewDecoder(req.Body).Decode(&observation)
	collection := client.Database(dbName).Collection(dbCollection)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	result, err := collection.InsertOne(ctx, observation)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		if _, err := resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`)); err != nil {
			//the blank field returns the number of bytes written
			log.Println(err.Error())
		}
	}

	if err := json.NewEncoder(resWriter).Encode(result); err != nil {
		log.Println(err.Error())
	}
}

func GetObservation(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	log.Println(id)
	filter := bson.M{"_id": id}

	var observation Observation

	collection := client.Database(dbName).Collection(dbCollection)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	err := collection.FindOne(ctx, filter).Decode(&observation)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		if _, err := resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`)); err != nil {
			log.Println(err.Error())
		}
	} else {
		resWriter.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(resWriter).Encode(observation); err != nil {
			log.Println(err.Error())
		}
	}
}

func main() {
	log.Println("Starting service.")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server "+
		"gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	clientOptions := options.Client().ApplyURI(dbAddress)
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()

	router.HandleFunc(versionEndpoint, GetVersion).Methods("GET")
	router.HandleFunc(observationEndpoint, CreateObservation).Methods("POST")
	router.HandleFunc(observationEndpoint+"/{id}", GetObservation).Methods("GET")

	srv := &http.Server{
		Addr:         "0.0.0.0:8088",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err.Error())
		}
	}()

	log.Println("Service started.")

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

	if err := srv.Shutdown(ctx); err != nil {
		log.Println(err.Error())
	}

	log.Println("Shutting down service.")

	os.Exit(0)
}
