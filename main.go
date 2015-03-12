package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/stretchr/graceful"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

func getURL(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	vars := mux.Vars(r)
	alias := vars["url"]
  var url *URLJson
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		v := b.Get([]byte(alias))
		if v != nil {
			err := incrementCount(alias, db)
			if err != nil {
				fmt.Println(err)
			}
      err = json.Unmarshal(v, &url)
      if err != nil {
        fmt.Println("error:", err)
      }
			http.Redirect(w, r, *url.Url, 302)
		} else {
			fmt.Fprintf(w, alias+" is not a valid alias.")
		}
		return nil
	})
}

func shortURL(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	newurl := string(r.FormValue("newurl"))
	urlvar := string(r.FormValue("url"))
	url, err := url.QueryUnescape(urlvar)
	if err != nil {
		fmt.Println(err)
	}
	var jsonr []byte

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		v := b.Get([]byte(newurl))
		if v != nil {
			error := Error{"URL", "Alias already exists."}
			errorjson := ErrorJson{error}
			jsonr, err = json.Marshal(errorjson)
			if err != nil {
				fmt.Println(err)
			}
		} else {
      count := 0
      urlst := URL{&url, &count}
			urljson := URLJson{urlst}
			jsonr, err = json.Marshal(urljson)
			if err != nil {
				fmt.Println(err)
        return nil
			}
			db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("urls"))
        return b.Put([]byte(newurl), jsonr)
			})
		}
		return nil
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonr))
}

func getAnalytics(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	vars := mux.Vars(r)
	var jsonr []byte
	url := vars["url"]
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		v := b.Get([]byte(url))
		if v != nil {
			jsonr = v
		} else {
			error := Error{"URL", vars["url"] + " is not a valid alias."}
			errorjson := ErrorJson{error}
			jsonr, err = json.Marshal(errorjson)
			if err != nil {
				fmt.Println(err)
			}
		}
		return nil
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonr))
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
	r.HandleFunc("/u/{url}", func(w http.ResponseWriter, r *http.Request) {
		getURL(w, r, db)
	})
	r.HandleFunc("/u/{url}/analytics", func(w http.ResponseWriter, r *http.Request) {
		getAnalytics(w, r, db)
	})
	r.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		shortURL(w, r, db)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rootHandler(w, r, db)
	})
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

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

func incrementCount(alias string, db *bolt.DB) (error) {
	var totalcount []byte
  var url *URLJson
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
    urljson := b.Get([]byte(alias))
    err := json.Unmarshal(urljson, &url)
    if err != nil {
      fmt.Println("error:", err)
    }
    count := *url.Count + 1
    *url.Count = count
    
    urljson, err = json.Marshal(&url)
    if err != nil {
      fmt.Println(err)
      return nil
    }
    err = b.Put([]byte(alias), urljson)

		b = tx.Bucket([]byte("stats"))
		totalcount = b.Get([]byte("totalcount"))
		if totalcount == nil {
			err = b.Put([]byte("totalcount"), []byte("0"))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			newcount, err := incStr((string(totalcount)))
			if err != nil {
				fmt.Println(err)
			}
			err = b.Put([]byte("totalcount"), []byte(newcount))
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})
	return nil
}

func incStr(count string) (string, error) {
	i, err := strconv.Atoi(count)
	if err != nil {
		return count, err
	}
	t := i + 1
	return strconv.Itoa(t), nil
}
