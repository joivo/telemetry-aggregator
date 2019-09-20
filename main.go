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

	"github.com/emanueljoivo/telemetry-aggregator/pusher"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DbAddrConfKey       = "MONGODB_ADDR"
	DbNameConfKey       = "MONGODB_DATABASE_NAME"
	DbCollectionConfKey = "MONGODB_DATABASE_COLLECTION_METRICS"

	MetricEndpoint     = "/metric"
	VersionEndpoint    = "/version"

	DefaultVersion     = "v1.0.0"
	DefaultTimeout     = 10 * time.Second
)

type Version struct {
	Tag string `json:"tag"`
}

var (
	client *mongo.Client
)

func GetVersion(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(resWriter).Encode(Version{DefaultVersion}); err != nil {
		log.Println(err.Error())
	}
}

func CreateMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	var observation pusher.Metric
	var p pusher.Pusher = pusher.PrometheusPusher{}

	_ = json.NewDecoder(req.Body).Decode(&observation)

	collection := client.Database(os.Getenv(DbNameConfKey)).Collection(os.Getenv(DbCollectionConfKey))

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)

	result, err := collection.InsertOne(ctx, &observation)

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		if _, err := resWriter.Write([]byte(`{ "message": "` + err.Error() + `" }`)); err != nil {
			// the blank field returns the number of bytes written
			log.Println(err.Error())
		}
	}

	p.PushMetric(&observation) // sends the data to the prometheus push gateway

	if err := json.NewEncoder(resWriter).Encode(result); err != nil {
		log.Println(err.Error())
	}
}

func GetMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	filter := bson.M{"_id": id}

	var observation pusher.Metric

	collection := client.Database(os.Getenv(DbNameConfKey)).Collection(os.Getenv(DbCollectionConfKey))
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
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

func validateEnv() {
	if _, exists := os.LookupEnv(DbAddrConfKey); !exists {
		log.Fatal("No database address on the environment.")
	} else if _ , exists := os.LookupEnv(DbNameConfKey); !exists {
		log.Fatal("No database name on the environment.")
	} else {
		log.Println("Environment loaded with success.")
	}
}

func init() {
	log.Println("Starting service.")

	validateEnv()}


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
