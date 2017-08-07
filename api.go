package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var APIClient = NewClient()

func NewClient() *Client {
	return &Client{
		client:  &http.Client{Timeout: 10 * time.Second},
		BaseURL: createAPIURL(),
	}
}

func createAPIURL() *url.URL {
	https := strings.ToLower(os.Getenv("TBS_HTTPS")) == "true"
	proto := "http"
	if https {
		proto = "https"
	}
	return &url.URL{Scheme: proto, Host: fmt.Sprintf("%s:%s", os.Getenv("TBS_HOST"), os.Getenv("TBS_PORT"))}

}

type Client struct {
	client  *http.Client
	BaseURL *url.URL
}

func (c *Client) NewRequest(method, urlStr, token string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return req, nil
}

func (c *Client) Get(urlStr, token string) (*http.Response, error) {
	req, err := c.NewRequest("GET", urlStr, token, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *Client) Post(urlStr, token string, body interface{}) (*http.Response, error) {
	req, err := c.NewRequest("POST", urlStr, token, body)
	if err != nil {
		return nil, err
	}
	log.Println("POST")
	return c.client.Do(req)
}
