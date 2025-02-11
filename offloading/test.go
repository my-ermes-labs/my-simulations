package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ermes-labs/api-go/api"
)

// Client struct to represent each client
type Client struct {
	ID           int
	Host         string
	SessionToken *api.SessionToken
}

// ResponseTime struct to store response time details
type ResponseTime struct {
	ClientID           int
	InitialRequestTime time.Time
	StartRequestTime   time.Time
	EndRequestTime     time.Time
}

// SliceData struct to store the data for each 400ms slice
type SliceData struct {
	AverageResponseTime float64
	Non200ResponseTime  float64
	Only200ResponseTime float64
}

var (
	responseTimes []ResponseTime
	mu            sync.Mutex
	wg            sync.WaitGroup
)

func (c *Client) ContinuousRequests(resource string, endTime time.Time) {
	defer wg.Done()
	for time.Now().Before(endTime) {
		startTime := time.Now()

		for {
			reqStartTime := time.Now()
			resp, err := c.makeRequest(resource)
			reqEndTime := time.Now()

			if err != nil {
				log.Printf("Client %d: Error making request: %v", c.ID, err)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				var startRequestTime time.Time
				if startRequestTime.IsZero() {
					startRequestTime = startTime
				} else {
					startRequestTime = reqStartTime
				}

				mu.Lock()
				responseTimes = append(responseTimes, ResponseTime{
					ClientID:           c.ID,
					InitialRequestTime: startTime,
					StartRequestTime:   startRequestTime,
					EndRequestTime:     reqEndTime,
				})
				mu.Unlock()
				break
			}
		}
	}
}

func (c *Client) makeRequest(resource string) (*http.Response, error) {
	url := fmt.Sprintf("http://%s/%s", c.Host, resource)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	tokenString, _ := api.MarshallSessionToken(*c.SessionToken)

	// Add the token to the request header if available
	if c.SessionToken != nil {
		req.Header.Set("X-Ermes-token", string(tokenString))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Extract the token from the header
	tokenString = []byte(resp.Header.Get("X-Ermes-token"))
	if len(tokenString) != 0 {
		token, err := api.UnmarshallSessionToken(tokenString)
		if err != nil {
			return nil, err
		}
		// Update the session token with the new one from the header
		c.SessionToken = token
		c.Host = token.Host
		log.Printf("Client %d: Host updated to %s", c.ID, c.Host)
	}

	return resp, nil
}

func test() {
	edgeNodeIP := "<edge-node-ip>"
	clients := make([]Client, 60)
	resources := []string{"/test/speech-to-text", "/test/cdn-upload", "/test/speech-to-text", "/test/cdn-download"}

	// Initialize clients
	for i := 0; i < 60; i++ {
		clients[i] = Client{
			ID:   i + 1,
			Host: edgeNodeIP,
		}
	}

	// Sequentially make requests to create sessions
	for i := 0; i < 60; i++ {
		clients[i].makeRequest("test/create-session")
	}

	// Wait for 3 seconds before the next steps
	time.Sleep(3 * time.Second)

	endTime := time.Now().Add(20 * time.Second)

	// Start continuous requests for the first client
	wg.Add(1)
	go clients[0].ContinuousRequests(resources[0], endTime)

	// Wait for 5 seconds before starting other clients
	time.Sleep(5 * time.Second)

	// Start continuous requests for other clients, staggered by 15ms
	for i := 1; i < 60; i++ {
		wg.Add(1)
		go clients[i].ContinuousRequests(resources[i%len(resources)], endTime)
		time.Sleep(15 * time.Millisecond)
	}

	wg.Wait()

	// Process response times into slices
	sliceData := processResponseTimes(responseTimes, 400*time.Millisecond)

	// Save slice data to a CSV file
	saveSliceData("slice_data.csv", sliceData)
}

func processResponseTimes(responseTimes []ResponseTime, sliceDuration time.Duration) []SliceData {
	var sliceData []SliceData
	startTime := responseTimes[0].InitialRequestTime
	endTime := startTime.Add(sliceDuration)

	for {
		var totalResponseTime float64
		var non200ResponseTime float64
		var count int

		for _, rt := range responseTimes {
			if rt.InitialRequestTime.Before(endTime) && rt.EndRequestTime.After(startTime) {
				overlapStart := maxTime(rt.InitialRequestTime, startTime)
				overlapEnd := minTime(rt.EndRequestTime, endTime)
				overlapDuration := overlapEnd.Sub(overlapStart).Seconds()
				weight := overlapDuration / sliceDuration.Seconds()

				totalResponseTime += weight * rt.EndRequestTime.Sub(rt.InitialRequestTime).Seconds()

				if rt.StartRequestTime != rt.InitialRequestTime && rt.StartRequestTime.Before(endTime) && rt.StartRequestTime.After(startTime) {
					non200OverlapStart := maxTime(rt.InitialRequestTime, startTime)
					non200OverlapEnd := minTime(rt.StartRequestTime, endTime)
					non200OverlapDuration := non200OverlapEnd.Sub(non200OverlapStart).Seconds()
					non200Weight := non200OverlapDuration / sliceDuration.Seconds()
					non200ResponseTime += non200Weight * rt.StartRequestTime.Sub(rt.InitialRequestTime).Seconds()
				}

				count++
			}
		}

		if count > 0 {
			only200ResponseTime := totalResponseTime - non200ResponseTime
			sliceData = append(sliceData, SliceData{
				AverageResponseTime: totalResponseTime / float64(count),
				Non200ResponseTime:  non200ResponseTime,
				Only200ResponseTime: only200ResponseTime,
			})
		}

		startTime = endTime
		endTime = endTime.Add(sliceDuration)

		if startTime.After(responseTimes[len(responseTimes)-1].EndRequestTime) {
			break
		}
	}

	return sliceData
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func saveSliceData(filename string, sliceData []SliceData) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Only200ResponseTime", "Non200ResponseTime"})

	for _, sd := range sliceData {
		writer.Write([]string{
			fmt.Sprintf("%.2f", sd.Only200ResponseTime),
			fmt.Sprintf("%.2f", sd.Non200ResponseTime),
		})
	}
}
