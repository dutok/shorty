package main

import (
	"net/http"
	"fmt"
  "os"
  "time"
  "encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/stretchr/graceful"
	"github.com/garyburd/redigo/redis"
)

var err error

type ErrorJson struct {
  Error `json:"error"`
}

type Error struct {
  Type  string `json:"type"`
  Message  string `json:"message"`
}

type URLJson struct {
  URL `json:"url"`
}
type URL struct {
  URL  string `json:"url"`
  NewURL  string `json:"newurl"`
}

func getURL(w http.ResponseWriter, r *http.Request, db redis.Conn) {
  vars := mux.Vars(r)
	var url string
  url, err := redis.String(db.Do("GET", vars["url"]))
  var b []byte
  if err != nil {
    error := Error{"URL", "URL does not exist"}
    errorjson := ErrorJson{error}
    b, err = json.Marshal(errorjson)
    if err != nil {
      fmt.Println(err)
      return
    }
  } else {
    urlst := URL{url, vars["url"]}
    urljson := URLJson{urlst}
    b, err = json.Marshal(urljson)
    if err != nil {
      fmt.Println(err)
      return
    }
  }
  w.Header().Set("Content-Type", "application/json")
  fmt.Fprintf(w, string(b))
}

func shortURL(w http.ResponseWriter, r *http.Request, db redis.Conn) {
  vars := mux.Vars(r)
  newurl := vars["newurl"]
  url := vars["url"]
	db.Do("SET", newurl, url)
  urlst := URL{url, newurl}
  urljson := URLJson{urlst}
  b, err := json.Marshal(urljson)
  if err != nil {
    fmt.Println(err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  fmt.Fprintf(w, string(b))
}

func main() {
    db, err := redis.Dial("tcp", ":"+os.Getenv("REDIS_PORT"))
    if err != nil {
      panic(err)
    }
    defer db.Close()
  
    r := mux.NewRouter()
    r.HandleFunc("/u/{url}", func(w http.ResponseWriter, r *http.Request) {
      getURL(w, r, db)
    })
    r.HandleFunc("/s/{url}/{newurl}", func(w http.ResponseWriter, r *http.Request){
      shortURL(w, r, db)
    })
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
 

    n := negroni.New()
    n.Use(gzip.Gzip(gzip.BestSpeed))

    n.UseHandler(r)

    graceful.Run(":"+os.Getenv("PORT"), 10*time.Second, n)
}