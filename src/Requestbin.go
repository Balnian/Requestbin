package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/teris-io/shortid"
)

const (
	//EntryLifeDuration Time an entry is kept before being deleted
	EntryLifeDuration = time.Hour * 4
	ListenPort        = ":8080"
	MaxBodyDataSize   = 1024 * 5
)

var srv http.Server

type reqdata struct {
	RemoteAddr string
	Method     string
	Header     http.Header
	URL        url.URL
	Body       []byte
	Time       time.Time
}

type dataEntry struct {
	Data         []reqdata
	CreationTime time.Time
}

var datastore map[string]dataEntry

func main() {
	datastore = make(map[string]dataEntry)
	go cleaner()
	http.HandleFunc("/", handleGeneric)
	http.HandleFunc("/new", handleNew)
	http.HandleFunc("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./html/js"))).ServeHTTP)
	http.HandleFunc("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./html/css"))).ServeHTTP)
	log.Fatal(http.ListenAndServe(ListenPort, nil))
}

func handleGeneric(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		handleRequest(w, req)
	} else {
		handleHome(w, req)
	}
}

func handleHome(w http.ResponseWriter, req *http.Request) {
	http.FileServer(http.Dir("./html/home")).ServeHTTP(w, req)

}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	id := strings.Trim(req.URL.Path, "/") // get id part of URL
	/*	if strings.Count(req.URL.Path, "/") >= 2 && req.URL.Query().Get("view") == "html" { // get id part of URL in case we are requesting files
		id = strings.SplitN(req.URL.Path, "/", 3)[1]
	}*/
	if _, ok := datastore[id]; ok {

		switch req.URL.Query().Get("view") {
		case "json":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(datastore[id].Data)
		case "html":
			prefix := "/" + id // Prefix to be remove in URL so we can find the file to serve
			/*if strings.Count(req.URL.Path, "/") >= 2 {
				prefix += "/"
			}*/
			http.StripPrefix(prefix, http.FileServer(http.Dir("./html/entry"))).ServeHTTP(w, req)
		default:
			// Get request body
			bdy := make([]byte, MaxBodyDataSize)
			nb, _ := req.Body.Read(bdy)
			//Get remote address, if theres a X-forwarded-for we take it over "req.RemoteAddr"
			remaddr := req.RemoteAddr
			if _, valid := req.Header["X-Forwarded-For"]; valid {
				remaddr = req.Header["X-Forwarded-For"][0]
			}
			datastore[id] = dataEntry{append(datastore[id].Data, reqdata{remaddr, req.Method, req.Header, *req.URL, bdy[:nb], time.Now()}), datastore[id].CreationTime}

		}
	} else {
		http.NotFound(w, req)
	}
}

func handleNew(w http.ResponseWriter, req *http.Request) {
	id, _ := shortid.Generate()
	datastore[id] = dataEntry{nil, time.Now()}
	http.Redirect(w, req, "/"+id+"?view=html", 303)
}

func cleaner() {
	for {
		for key, value := range datastore {
			if time.Now().Sub(value.CreationTime) >= EntryLifeDuration {
				delete(datastore, key)
			}
			time.Sleep(time.Hour)
		}
	}
}
