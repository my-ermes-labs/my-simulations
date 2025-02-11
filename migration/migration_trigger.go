package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	// Create a new HTTP client
	client := &http.Client{}
	// Session size
	sessionSize := r.URL.Query().Get("size")

	// Create a URL with the query parameters
	url, err := url.Parse("http://localhost/ermes-api/migration")
	if err != nil {
		log.Fatal(err)
	}
	query := url.Query()
	query.Set("session-id", sessionSize)
	query.Set("node-area", "central-usa")
	url.RawQuery = query.Encode()

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Start the timer
	start := time.Now()

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Calculate the elapsed time
	elapsed := time.Since(start)

	// Write the response
	io.WriteString(w, strconv.FormatInt(elapsed.Milliseconds(), 10)+"ms")
}
