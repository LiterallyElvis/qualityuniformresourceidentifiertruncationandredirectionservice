package main

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var db *bolt.DB
var primaryBucketName []byte

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
	log.Printf("r: %v", string(r))
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

func apiAddValue(w http.ResponseWriter, r *http.Request) {
	log.Println("apiAddValue called")
	vars := mux.Vars(r)
	k := []byte(vars["key"])
	v := []byte(vars["value"])
	addKeyValueToBucket(primaryBucketName, k, v)
}

func apiGetValue(w http.ResponseWriter, r *http.Request) {
	log.Println("apiGetValue called")
	vars := mux.Vars(r)
	k := []byte(vars["key"])
	rv := readKeyFromBucket(primaryBucketName, k)
	w.Write(rv)
}

func notFoundError(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Sorry, 404! :("))
}

func main() {
	primaryBucketName = []byte("Links")
	var err error
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err = bolt.Open("qualityuniformresourceidentifiertruncationandredirectionservice.db", 0600, nil)
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
	router.Methods("GET", "POST")

	router.HandleFunc("/api/add/{key}/{value}", apiAddValue)
	router.HandleFunc("/api/get/{key}", apiGetValue)
	// router.NotFoundHandler = http.HandlerFunc(notFoundError)

	http.Handle("/api", router)
	http.ListenAndServe(":3000", nil)
}
