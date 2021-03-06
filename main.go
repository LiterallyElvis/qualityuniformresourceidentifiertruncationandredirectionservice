package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var db *bolt.DB
var primaryBucketName []byte
var markov *Chain
var domain string

func init() {
	markov = buildMarkov()
	primaryBucketName = []byte("Links")
	domain = os.Getenv("DOMAIN")
}

type shortenerResponse struct {
	Key         string `json:",omitempty"`
	Destination string `json:",omitempty"`
	Error       string `json:",omitempty"`
}

func returnJSONFromStruct(s shortenerResponse) []byte {
	o, err := json.Marshal(s)
	if err != nil {
		log.Printf("error occurred marshalling notFoundError response: %v\n", err)
	}
	return o
}

func createBucket(name []byte) error {
	// Start a writable transaction.
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucket(name)
	if err != nil {
		log.Println("Bucket already exists!")
	}

	// Commit the transaction and check for error.
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func readKeyFromBucket(bucketName []byte, key []byte) []byte {
	var r []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		r = b.Get(key)
		return nil
	})
	return r
}

func addKeyValueToBucket(bucketName []byte, key []byte, value []byte) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		err := b.Put(key, value)
		if err != nil {
			log.Printf("there was an error:\n\t%v", err)
		}
		return err
	})
}

// prints a list of all keys. Not used in any routes, but handy to keep around, should the need arise.
func keys() {
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(primaryBucketName)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
}

func apiAddValue(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	if strings.Contains(url, domain) {
		// If the url provided contains the domain name we're running, we respond with an error
		respondWithError(w, r, "Invalid URL Provided! You musn't shorten links to the service itself!")
		return
	}

	if len(url) == 0 {
		// If no url query param is provided, we respond with an error
		respondWithError(w, r, "Empty URL Provided!")
		return
	}

	var result shortenerResponse
	var v []byte
	m := []byte(" ")

	for len(m) != 0 {
		v = generateMarkovString(markov)
		m = readKeyFromBucket(primaryBucketName, v)
	}

	result.Key = string(v)
	o, err := json.Marshal(result)
	if err != nil {
		log.Printf("error encountered marshalling json: %v\n", err)
	}

	addKeyValueToBucket(primaryBucketName, v, []byte(url))
	w.Write(o)
}

func apiGetValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k := url.QueryEscape(vars["key"])
	if len(k) != 0 {
		rv := readKeyFromBucket(primaryBucketName, []byte(k))
		http.Redirect(w, r, string(rv), 301)
	} else {
		notFoundError(w, r)
	}
}

func respondWithError(w http.ResponseWriter, r *http.Request, e string) {
	w.Write(returnJSONFromStruct(
		shortenerResponse{
			Error: e,
		}))
}

func homepage(w http.ResponseWriter, r *http.Request) {
	indexPage, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Printf("error occurred reading indexPage: %v\n", err)
	}
	w.Write(indexPage)
}

func notFoundError(w http.ResponseWriter, r *http.Request) {
	response := shortenerResponse{
		Error: "Sorry, 404! :(",
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(returnJSONFromStruct(response))
}

func main() {
	var err error
	// Open the db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err = bolt.Open("whatever.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ensure our default bucket exists
	err = createBucket(primaryBucketName)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", homepage)
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	// I know I need to actually serve the above files, but
	// I don't feel like it at the moment, sue me if you must.

	router.HandleFunc("/api/add", apiAddValue)
	router.HandleFunc("/{key}", apiGetValue)
	router.NotFoundHandler = http.HandlerFunc(notFoundError)

	http.Handle("/", router)
	http.ListenAndServe(":80", nil)
}
