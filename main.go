package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
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
	DbCollectionConfKey = "MONGODB_DATABASE_COLLECTION_AGGREGATOR"
	DbUsernameConfKey = "MONGODB_USERNAME"
	DbPasswdConfKey = "MONGODB_PASSWD"

	MetricEndpoint     = "/metric"
	VersionEndpoint    = "/version"

	DefaultVersion     = "v1.0.0"
	DefaultTimeout     = 10 * time.Second
)

type Version struct {
	Tag string `json:"tag"`
}

var (
	db   *mongo.Database
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

	dbCollection, collExists := os.LookupEnv(DbCollectionConfKey)

	if collExists {
			collection := db.Collection(dbCollection)
			ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
			result, err := collection.InsertOne(ctx, observation)

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
	} else {
		log.Println("No collection with the name specified on the environment.")
	}
}

func GetMetric(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("content-type", "application/json")

	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	log.Println(id)
	filter := bson.M{"_id": id}

	var observation pusher.Metric

	dbCollection, collExists := os.LookupEnv(DbCollectionConfKey)
	
	if collExists {
			collection := db.Collection(dbCollection)
			ctx, _ := context.WithTimeout(context.Background(),DefaultTimeout)
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

	} else {
		log.Println("No collection with the name specified on the environment.")
	}	
	
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found.")
	}
}

func configDB(ctx context.Context) (*mongo.Database, error) {

	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		ctx.Value(DbUsernameConfKey).(string),
		ctx.Value(DbPasswdConfKey).(string),
		ctx.Value(DbAddrConfKey).(string),
		ctx.Value(DbNameConfKey).(string),
	)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("todo: couldn't connect to mongo: %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("todo: mongo db couldn't connect with background context: %v", err)
	}
	aggregatordb := client.Database("todo")
	return aggregatordb, nil
}

func validateEnv() {
	if _, exists := os.LookupEnv(DbAddrConfKey); !exists {
		log.Fatal("No database address on the environment.")
	} else if _ , exists := os.LookupEnv(DbNameConfKey); !exists {
		log.Fatal("No database name on the environment.")
	} else if _ , exists := os.LookupEnv(DbUsernameConfKey); !exists {
		log.Fatal("No user name on the environment.")
	} else if _ , exists := os.LookupEnv(DbPasswdConfKey); !exists {
		log.Fatal("No user password on the environment.")
	} else {
		log.Println("Environment loaded with success.")
	}
}

func main() {
	log.Println("Starting service.")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server "+
		"gracefully wait for existing connections to finish - e.g. 15s or 1m")

	var addr  = flag.String("listen-address", ":8088", "The address to listen on for HTTP requests.")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	validateEnv()

	ctx = context.WithValue(ctx, DbAddrConfKey, os.Getenv(DbAddrConfKey))
	ctx = context.WithValue(ctx, DbNameConfKey, os.Getenv(DbNameConfKey))
	ctx = context.WithValue(ctx, DbUsernameConfKey, os.Getenv(DbUsernameConfKey))
	ctx = context.WithValue(ctx, DbPasswdConfKey, os.Getenv(DbPasswdConfKey))

	db, _ = configDB(ctx)

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
