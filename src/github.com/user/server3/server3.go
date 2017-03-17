package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"sync"
	"io/ioutil"
)

var backendURL *url.URL
var mutex = &sync.Mutex{}

func main() {
	var port uint
	var backend string
	var maxConcurrency uint
	var timeout time.Duration
	var err error
	var output string


	flag.UintVar(&port, "port", 80, "listening port")
	flag.StringVar(&backend, "backend", "http://localhost/", "backend http service url")
	flag.DurationVar(&timeout, "timeout", 300*time.Millisecond, "request timeout deadline")
	flag.UintVar(&maxConcurrency, "concurrency", 4, "max concurrency for backend requests")

	flag.Parse()

	backendURL, err = url.Parse(backend)
	if err != nil {
		log.Fatal("url Parse: ", err)
	}

	var mux = http.NewServeMux() //Creating HTTP Multiplexer to create a new Handler
	var gate = make(chan struct{}, maxConcurrency) //Creating a buffer channel of size maxConcurrency

	mux.HandleFunc("/",func (w http.ResponseWriter, r *http.Request) {
		//Using http.Get to get response from the backendURL
		response, err := http.Get(backendURL.String())
		if err != nil {
			log.Fatal(err)
		}
		body, _ := ioutil.ReadAll(response.Body)
		output = string(body)

		gate <- struct{}{}
		defer func() { <-gate }()

		//Fetch backendURL and return contents to the Client
    		fmt.Fprintf(w, output)

		//time.Sleep(6 * time.Second) //For testing of the code
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), http.TimeoutHandler(mux, timeout, "504 Server Response Timeout error"))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
