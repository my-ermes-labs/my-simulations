package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ResponseData struct {
	Time       int64
	StatusCode int
	TotalTime  int64
}

func sendRequest(url string) (ResponseData, string, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return ResponseData{}, "", err
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()
	if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound {
		newURL := resp.Header.Get("Location")
		if newURL != "" {
			startRedirect := time.Now()
			respRedirect, err := http.Get(newURL)
			if err != nil {
				return ResponseData{}, "", err
			}
			defer respRedirect.Body.Close()
			durationRedirect := time.Since(startRedirect).Milliseconds()
			return ResponseData{Time: time.Now().Unix(), StatusCode: resp.StatusCode, TotalTime: duration + durationRedirect}, newURL, nil
		}
	}

	return ResponseData{Time: time.Now().Unix(), StatusCode: resp.StatusCode, TotalTime: duration}, "", nil
}

func writeCSV(filename string, data [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	url := "http://192.168.64.25/api/endpoint"

	var data [][]string
	data = append(data, []string{"Time", "ResponseCode", "ResponseTime"})

	for i := 0; i < 30; i++ {
		respData, newIp, err := sendRequest(url)
		if err != nil {
			fmt.Println("Error sending request:", err)
			continue
		}
		data = append(data, []string{strconv.FormatInt(respData.Time, 10), strconv.Itoa(respData.StatusCode), strconv.FormatInt(respData.TotalTime, 10)})
		if newIp != "" {
			url = newIp + "/api/endpoint"
		}
	}

	if err := writeCSV("responses.csv", data); err != nil {
		fmt.Println("Error writing CSV:", err)
	}
}
