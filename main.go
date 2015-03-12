package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/stretchr/graceful"
	"html/template"
	"net/http"
	"os"
	"time"
)

type Stats struct {
	UrlCount   *int
	TotalCount *string
}

var err error

type ErrorJson struct {
	Error `json:"error"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type URLJson struct {
	URL `json:"url"`
}
type URL struct {
	Url   *string `json:"url"`
	Count *int    `json:"count"`
}

func main() {
	var file string
	if os.Getenv("DEV") == "true" {
		file = "data/my.db"
	} else {
		file = "/data/my.db"
	}
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("stats"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	r := mux.NewRouter()
	loadRoutes(r, db)

	n := negroni.New()
	n.Use(gzip.Gzip(gzip.BestSpeed))

	n.UseHandler(r)

	graceful.Run(":"+os.Getenv("PORT"), 10*time.Second, n)
}

func rootHandler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	t, _ := template.ParseFiles("./public/index.html")
	var urlcount int
	var totalcount []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		stats := b.Stats()
		urlcount = stats.KeyN / 2

		b = tx.Bucket([]byte("stats"))
		totalcount = b.Get([]byte("totalcount"))
		return nil
	})
	newtotalcount := string(totalcount)
	stats := Stats{&urlcount, &newtotalcount}
	t.Execute(w, stats)
}
