package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type ResourceResult struct {
	Success bool
	Error   error
	Body    string
	ETag    string
}

func post(url string, body string) ResourceResult {
	return modifyRequest(http.MethodPost, url, getAuthenticator(), bytes.NewBuffer([]byte(body)), "")
}

func postStream(url string, body io.Reader) ResourceResult {
	return modifyRequest(http.MethodPost, url, getAuthenticator(), body, "")
}

func get(url string) ResourceResult {
	return queryRequest(getAuthenticator(), url)
}

func httpDelete(url string) ResourceResult {
	return modifyRequest(http.MethodDelete, url, getAuthenticator(), bytes.NewBuffer([]byte("")), "")
}

type RequestAuthenticator func(req *http.Request)

func AddCookieAuth(value string) RequestAuthenticator {
	return func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: "__Host-crude-session-token", Value: value})
	}
}

func AddAPIKeyAuth(apiKey string) RequestAuthenticator {
	return func(req *http.Request) {
		req.Header.Set("X-Crude-Api-Key", apiKey)
	}
}

func getAuthenticator() RequestAuthenticator {
	if sessionToken != "" {
		return AddCookieAuth(sessionToken)
	} else {
		return AddAPIKeyAuth(apiKey)
	}
}

func queryRequest(auth RequestAuthenticator, url string) ResourceResult {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	request.Header.Set("Content-Type", "application/json")
	auth(request)
	return executeRequest(request)
}

func modifyRequest(httpMethod string, url string, auth RequestAuthenticator, body io.Reader, etag string) ResourceResult {
	request, err := prepareWriteRequest(httpMethod, url, body, etag)
	if err != nil {
		return ResourceResult{Success: false, Error: err}
	}
	auth(request)
	return executeRequest(request)
}

func executeRequest(request *http.Request) ResourceResult {
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return ResourceResult{Success: false, Error: err}
	}
	defer resp.Body.Close()
	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body: " + err.Error())
		return ResourceResult{Success: false, Error: err}
	}
	newETag := resp.Header.Get("ETag")
	return ResourceResult{Success: true, Body: string(buffer), ETag: newETag}
}

func prepareWriteRequest(httpMethod string, url string, body io.Reader, etag string) (*http.Request, error) {
	request, err := http.NewRequest(httpMethod, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	if etag != "" {
		request.Header.Set("If-Match", etag)
		fmt.Println("Update at " + endPoint + " using etag: " + etag)
	}
	return request, nil
}
