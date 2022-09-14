package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	jsonc "github.com/nwidger/jsoncolor"
	"io"
	"log"
	"net/http"
)

type ResourceResult struct {
	Success bool
	Error   error
	Body    string
	ETag    string
}

func post(url string, body string) ResourceResult {
	return modifyRequest(http.MethodPost, url, apiKey, body, "")
}

func postUpdate(url string, body string, etag string) ResourceResult {
	return modifyRequest(http.MethodPost, url, apiKey, body, etag)
}

func get(url string) ResourceResult {
	return queryRequest(apiKey, url)
}

func httpDelete(url string) ResourceResult {
	return modifyRequest(http.MethodDelete, url, apiKey, "", "")
}

func queryRequest(apiKey string, url string) ResourceResult {
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Crude-Api-Key", apiKey)
	resp, err := client.Do(request)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	defer resp.Body.Close()
	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	etag := resp.Header.Get("ETag")
	return ResourceResult{Success: true, Body: string(buffer), ETag: etag}
}

func modifyRequest(httpMethod string, url string, apiKey string, body string, etag string) ResourceResult {
	client := &http.Client{}
	request, err := http.NewRequest(httpMethod, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Crude-Api-Key", apiKey)
	if etag != "" {
		request.Header.Set("If-Match", etag)
		fmt.Println("Update at " + endPoint + " using etag: " + etag)
	}
	resp, err := client.Do(request)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	defer resp.Body.Close()
	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	newETag := resp.Header.Get("ETag")
	return ResourceResult{Success: true, Body: string(buffer), ETag: newETag}
}

func parseJson(jsonString string) JsonObject {
	var jsonParsed JsonObject
	json.Unmarshal([]byte(jsonString), &jsonParsed)
	return jsonParsed
}

func toJson(input any) string {
	jsonString, err := json.Marshal(input)
	if err != nil {
		log.Fatalln(err)
	}
	return string(jsonString)
}

func jsonPrettify(input string) string {
	encoder := jsonc.NewFormatter()
	encoder.Indent = "  "
	var buffer bytes.Buffer
	if err := encoder.Format(&buffer, []byte(input)); err != nil {
		panic(err)
	}
	return buffer.String()
}
