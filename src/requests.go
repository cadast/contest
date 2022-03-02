package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type RequestResult struct {
	StatusCode   int
	Body         []byte
	ContentType  string
	ResponseTime int64
}

func RunRequest(url string, headers map[string]string, body []byte) (*RequestResult, error) {
	client := http.Client{}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodGet, url, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	rres, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return nil, err
	}

	if rres.Body != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(rres.Body)
	}

	res := RequestResult{
		StatusCode:   rres.StatusCode,
		ResponseTime: responseTime.Milliseconds(),
	}

	responseBody, err := ioutil.ReadAll(rres.Body)
	if err != nil {
		return nil, err
	}
	res.Body = responseBody
	res.ContentType = rres.Header.Get("Content-Type")

	return &res, nil
}
