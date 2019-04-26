package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
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
var client appinsights.TelemetryClient

func main() {
	fmt.Println("Creating TelemetryClient")
	client = appinsights.NewTelemetryClient("")

	client.TrackEvent("Server Starting")
	defer client.TrackEvent("Server Shutting Down")

	fmt.Println("Creating Database")
	datastore = make(map[string]dataEntry)

	fmt.Println("Starting Cleaner")
	go cleaner()

	fmt.Println("Starting Web Server")
	http.HandleFunc("/", handleGeneric)
	http.HandleFunc("/new", handleNew)
	http.HandleFunc("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./html/js"))).ServeHTTP)
	http.HandleFunc("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./html/css"))).ServeHTTP)
	http.HandleFunc("/AvailabilityTest", handleAvailabilityTest)
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
	reqBeginTime := time.Now()
	//Get remote address, if theres a X-forwarded-for we take it over "req.RemoteAddr"
	remaddr := req.RemoteAddr
	if _, valid := req.Header["X-Forwarded-For"]; valid {
		remaddr = req.Header["X-Forwarded-For"][0]
	}
	request := appinsights.NewRequestTelemetry(req.Method, req.URL.String(), 0, "200")
	request.Source = remaddr
	if _, ok := datastore[id]; ok {

		request.Success = true

		switch req.URL.Query().Get("view") {
		case "json":
			request.Properties["Type"] = "json"

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(datastore[id].Data)
		case "html":
			request.Properties["Type"] = "html"

			prefix := "/" + id // Prefix to be remove in URL so we can find the file to serve
			/*if strings.Count(req.URL.Path, "/") >= 2 {
				prefix += "/"
			}*/
			http.StripPrefix(prefix, http.FileServer(http.Dir("./html/entry"))).ServeHTTP(w, req)
		default:
			// Get request body
			bdy := make([]byte, MaxBodyDataSize)
			nb, _ := req.Body.Read(bdy)

			datastore[id] = dataEntry{append(datastore[id].Data, reqdata{remaddr, req.Method, req.Header, *req.URL, bdy[:nb], time.Now()}), datastore[id].CreationTime}
			//fmt.Println(string(bdy[:nb]))
			request.Properties["Type"] = "Logging"
			//request.Properties["Body Size"] = nb

			if nb > 0 {
				//request.Properties["Body"] = string(bdy[:nb])
			}
			w.WriteHeader(200)
		}
	} else {
		request.Success = true

		http.NotFound(w, req)
	}
	request.MarkTime(reqBeginTime, time.Now())
	client.Track(request)
}

func handleNew(w http.ResponseWriter, req *http.Request) {
	id, _ := shortid.Generate()
	datastore[id] = dataEntry{nil, time.Now()}
	http.Redirect(w, req, "/"+id+"?view=html", 303)

	event := appinsights.NewEventTelemetry("Created Bin")
	event.Properties["id"] = id
	event.SetTime(datastore[id].CreationTime)
	client.Track(event)

}

func handleAvailabilityTest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)

}

func cleaner() {
	for {
		time.Sleep(time.Hour)

		for key, value := range datastore {
			if time.Now().Sub(value.CreationTime) >= EntryLifeDuration {
				delete(datastore, key)

				event := appinsights.NewEventTelemetry("Deleted Bin")
				event.Properties["id"] = key
				client.Track(event)
			}
		}
		client.TrackMetric("Bin Number", float64(len(datastore)))
	}
}
