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
	HTTPClient *http.Client
}

const notifyApiUrl = "https://notify-api.line.me/api/notify"

// NewClient returns *Client
func NewClient() *Client {
	return &Client{HTTPClient: http.DefaultClient}
}

// NotifyMessage notify text message
func (c *Client) NotifyMessage(ctx context.Context, token string, message string) (*Response, error) {
	body, contentType, err := c.requestBody(message)
	if err != nil {
		return nil, err
	}
	return c.notify(ctx, token, body, contentType)
}

func (c *Client) notify(ctx context.Context, token string, body io.Reader, contentType string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", notifyApiUrl, body)
	if err != nil {
		return nil, fmt.Errorf("failed to new request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HTTPClient.Do(req)
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

func (c *Client) requestBody(message string) (io.Reader, string, error) {
	v := url.Values{}
	v.Add("message", message)
	return strings.NewReader(v.Encode()), "application/x-www-form-urlencoded", nil
}
