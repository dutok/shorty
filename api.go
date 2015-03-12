package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"net/url"
)

func loadRoutes(r *mux.Router, db *bolt.DB, shorty Shorty) {
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
		rootHandler(w, r, db, shorty)
	})
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
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
      http.Redirect(w, r, "/#404", 302)
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

func rootHandler(w http.ResponseWriter, r *http.Request, db *bolt.DB, shorty Shorty) {
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
	stats := Stats{&urlcount, &newtotalcount, shorty}
	t.Execute(w, stats)
}
