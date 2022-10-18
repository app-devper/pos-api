package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type Client struct {
	httpClient *http.Client
}

type configuration struct {
	token       string
	method      string
	body        io.Reader
	endpoint    string
	contentType string
}

func NewClient() *Client {
	return &Client{httpClient: http.DefaultClient}
}

func (c *Client) Notify(ctx context.Context, token string, message string) (*Response, error) {
	configuration := c.createConfiguration(token, message)
	req, err := http.NewRequestWithContext(ctx, configuration.method, configuration.endpoint, configuration.body)
	if err != nil {
		return nil, fmt.Errorf("failed to new request: %w", err)
	}
	req.Header.Set("Content-Type", configuration.contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", configuration.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to notify: %w", err)
	}
	nResp := &Response{}
	err = json.NewDecoder(resp.Body).Decode(nResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode notification response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nResp, errors.New("invalid access token")
	}

	if resp.StatusCode != http.StatusOK {
		return nResp, errors.New(nResp.Message)
	}
	return nResp, nil
}

func (c *Client) createConfiguration(token, message string) configuration {
	v := url.Values{}
	v.Add("message", message)
	return configuration{
		token:       token,
		endpoint:    "https://notify-api.line.me/api/notify",
		method:      "POST",
		body:        strings.NewReader(v.Encode()),
		contentType: "application/x-www-form-urlencoded",
	}
}
