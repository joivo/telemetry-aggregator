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

	"github.com/emanueljoivo/telemetry-aggregator/api"
	"github.com/emanueljoivo/telemetry-aggregator/pusher"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func GetVersion(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(resWriter).Encode(api.Version{api.DefaultVersion}); err != nil {
		log.Println(err.Error())
	}
}

func CreateMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	var observation pusher.Metric
	var p pusher.Pusher = pusher.PrometheusPusher{}

	_ = json.NewDecoder(req.Body).Decode(&observation)
	collection := client.Database(api.DatabaseName).Collection(api.DatabaseCollection)
	ctx, _ := context.WithTimeout(context.Background(), api.DefaultTimeout)
	result, err := collection.InsertOne(ctx, observation)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		if _, err := resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`)); err != nil {
			// the blank field returns the number of bytes written
			log.Println(err.Error())
		}
	}

	p.Push(&observation) // sends the data to the prometheus push gateway

	if err := json.NewEncoder(resWriter).Encode(result); err != nil {
		log.Println(err.Error())
	}
}

func GetMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	log.Println(id)
	filter := bson.M{"_id": id}

	var observation pusher.Metric

	collection := client.Database(api.DatabaseName).Collection(api.DatabaseCollection)
	ctx, _ := context.WithTimeout(context.Background(), api.DefaultTimeout)
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

	clientOptions := options.Client().ApplyURI(api.DatabaseAddr)
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()

	router.HandleFunc(api.VersionEndpoint, GetVersion).Methods("GET")
	router.HandleFunc(api.MetricEndpoint, CreateMetric).Methods("POST")
	router.HandleFunc(api.MetricEndpoint+"/{id}", GetMetric).Methods("GET")

	srv := &http.Server{
		Addr:         ":8088",
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
