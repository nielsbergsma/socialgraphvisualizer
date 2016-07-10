package main

import (
	"log"
	"fmt"
	"time"
	"flag"
	"net/http"
	"encoding/json"
	"encoding/base64"

	"./graph"
	"./model"
	"./crawler"

	"github.com/gorilla/mux"
	"github.com/ChimeraCoder/anaconda"
)

var (
	storage *graph.Graph
	twitterClient *anaconda.TwitterApi
	httpBindAddress = ":7999" 
)

func getGraph(w http.ResponseWriter, r *http.Request) {
	parameters := mux.Vars(r)

	actorUri, err := base64.StdEncoding.DecodeString(parameters["actor"])
	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{ "error": "Invalid actor uri (must be base64 encoded)" }`))
		return
	}

	graph, err := storage.GetFriendships(string(actorUri))
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err.Error())))
		return
	}

	body, _ := json.Marshal(graph)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(body)
}

func createGraph(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	body := struct {
		Uri string `json:"uri"`
	}{}

	if err := decoder.Decode(&body); err != nil || body.Uri == "" {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{ "error": "invalid body" }`))
	}

	crawler := crawler.NewTwitterCrawler(&model.Actor { Uri: body.Uri }, twitterClient)
	defer crawler.Close()
	go crawler.Run()

	for {
		select {
		case actor := <- crawler.Actor:
			log.Printf("Found actor %s (%s)", actor.Name, actor.Uri)
			storage.SetActor(actor, time.Now().UnixNano())
		
		case friendship := <- crawler.Friendship:		
			log.Printf("Found friendship %s (%s) - %s (%s)", friendship.From.Name, friendship.From.Uri, friendship.To.Name, friendship.To.Uri)
			storage.SetFriendship(friendship)

		case <- crawler.Completed:		
			break
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(`{ "success": true }`))	
}

func main() {
	neo4jConnectionString := flag.String("neo4j", "http://localhost:7474/db/data", "Neo4J connection string")
	twitterConsumerKey := flag.String("twitter-consumer-key", "", "Twitter consumer key")
	twitterConsumerSecret := flag.String("twitter-consumer-secret", "", "Twitter consumer secret")
	twitterAccessToken := flag.String("twitter-accesstoken", "", "Twitter access token")
	twitterAccessTokenSecret := flag.String("twitter-accesstoken-secret", "", "Twitter access token secret")
	flag.Parse()

	var err error
	storage, err = graph.Connect(*neo4jConnectionString)
	if err != nil {
		log.Fatalf("Unable to connect to social graph %+v", err)
		return
	}
	log.Printf("Connected to Neo4J")

	anaconda.SetConsumerKey(*twitterConsumerKey)
	anaconda.SetConsumerSecret(*twitterConsumerSecret)
	twitterClient = anaconda.NewTwitterApi(*twitterAccessToken, *twitterAccessTokenSecret)
	log.Printf("Connected to Twitter")

	router := mux.NewRouter()
	router.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("html/"))))
	router.HandleFunc("/graphs", createGraph).Methods("POST")
	router.HandleFunc("/graphs/{actor}", getGraph).Methods("Get")

	log.Printf("Http endpoint started on %s", httpBindAddress)
	http.Handle("/", router)

	if err := http.ListenAndServe(httpBindAddress, nil); err != nil {
		log.Fatalf("Unable to connect to social graph %+v", err)
		return
	}
}