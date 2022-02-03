package main

import (
	"io/ioutil"
	"net/http"
)

type RequestResult struct {
	StatusCode int
	Body       string
}

func RunRequest(url string, headers map[string]string) (*RequestResult, error) {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rres, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if rres.Body != nil {
		defer rres.Body.Close()
	}

	res := RequestResult{StatusCode: rres.StatusCode}

	body, err := ioutil.ReadAll(rres.Body)
	if err != nil {
		return nil, err
	}
	res.Body = string(body)

	return &res, nil
}
