package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/emanueljoivo/telemetry-aggregator/pkg/broker"
	"github.com/emanueljoivo/telemetry-aggregator/pkg/models"
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

const DbAddrConfKey       = "MONGODB_ADDR"
const DbNameConfKey       = "MONGODB_DATABASE_NAME"
const DbCollectionConfKey = "MONGODB_DATABASE_COLLECTION_METRICS"

const MetricEndpoint  = "/metric"
const VersionEndpoint = "/version"

const DefaultVersion = "v1.1.0"

var client *mongo.Client


func GetVersion(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(resWriter).Encode(models.Version{Tag: DefaultVersion}); err != nil {
		log.Println(err.Error())
	}
}

func CreateMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	var metric models.Metric

	_ = json.NewDecoder(req.Body).Decode(&metric)

	collection := client.Database(os.Getenv(DbNameConfKey)).Collection(os.Getenv(DbCollectionConfKey))

	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)

	result, err := collection.InsertOne(ctx, &metric)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		if _, err := resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`)); err != nil {
			// the blank field returns the number of bytes written
			log.Println(err.Error())
		}
	}

	broker.Place(&metric) // place the metric in a pool

	if err := json.NewEncoder(resWriter).Encode(result); err != nil {
		log.Println(err.Error())
	}
}

func GetMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	filter := bson.M{"_id": id}

	var metric models.Metric

	collection := client.Database(os.Getenv(DbNameConfKey)).Collection(os.Getenv(DbCollectionConfKey))
	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
	err := collection.FindOne(ctx, filter).Decode(&metric)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		if _, err := resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`)); err != nil {
			log.Println(err.Error())
		}
	} else {
		resWriter.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(resWriter).Encode(metric); err != nil {
			log.Println(err.Error())
		}
	}
}

func validateEnv() {
	if _, exists := os.LookupEnv(DbAddrConfKey); !exists {
		log.Fatal("No database address on the environment.")
	} else if _, exists := os.LookupEnv(DbNameConfKey); !exists {
		log.Fatal("No database name on the environment.")
	} else {
		log.Println("Environment loaded with success.")
	}
}

func init() {
	log.Println("Starting service.")

	validateEnv()
}

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server "+
		"gracefully wait for existing connections to finish - e.g. 15s or 1m")
	var addr = flag.String("listen-address", ":8088", "The address to listen on for HTTP requests.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	clientOptions := options.Client().ApplyURI(os.Getenv(DbAddrConfKey))
	client, _ = mongo.Connect(ctx, clientOptions)

	err := client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal("Error to connect with db: ", err)
	}

	log.Println("Connected with the database.")

	broker.StartBroker()

	router := mux.NewRouter()

	router.HandleFunc(VersionEndpoint, GetVersion).Methods("GET")
	router.HandleFunc(MetricEndpoint, CreateMetric).Methods("POST")
	router.HandleFunc(MetricEndpoint+"/{id}", GetMetric).Methods("GET")

	srv := &http.Server{
		Addr:         *addr,
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
