package mailcow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	ConstAPIKeyHeaderName = "X-API-Key" //nolint:gosec
)

type client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func New(
	httpClient *http.Client,
	baseURL,
	apiKey string,
) Client {
	return &client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

type Client interface {
	GetMailboxes(ctx context.Context) ([]Mailbox, error)
	GetAppPasswords(ctx context.Context, username string) ([]AppPassword, error)
	DeleteAppPasswords(ctx context.Context, ids []int) error
	CreateAppPassword(ctx context.Context, username, appname, password, protocols string) error
}

func (c *client) GetMailboxes(ctx context.Context) ([]Mailbox, error) {
	reqURL, err := url.JoinPath(c.baseURL, "api/v1/get/mailbox/all")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(ConstAPIKeyHeaderName, c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	mb := make([]Mailbox, 0)
	if err := json.NewDecoder(resp.Body).Decode(&mb); err != nil {
		return nil, err
	}
	return mb, nil
}

func (c *client) GetAppPasswords(ctx context.Context, username string) ([]AppPassword, error) {
	reqURL, err := url.JoinPath(c.baseURL, "/api/v1/get/app-passwd/all", username)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(ConstAPIKeyHeaderName, c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	ap := make([]AppPassword, 0)
	if err := json.NewDecoder(resp.Body).Decode(&ap); err != nil {
		if strings.HasPrefix(err.Error(), "json: cannot unmarshal object into Go value of type []") {
			return []AppPassword{}, nil
		}
		return nil, err
	}
	return ap, nil
}

func (c *client) DeleteAppPasswords(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	reqURL, err := url.JoinPath(c.baseURL, "api/v1/delete/app-passwd")
	if err != nil {
		return err
	}
	reqBytes, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	req.Header.Add(ConstAPIKeyHeaderName, c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (c *client) CreateAppPassword(ctx context.Context, username, appname, password, protocols string) error {
	reqURL, err := url.JoinPath(c.baseURL, "api/v1/add/app-passwd")
	if err != nil {
		return err
	}
	reqStruct := struct {
		Username  string `json:"username"`
		AppName   string `json:"app_name"`
		Password  string `json:"app_passwd"`
		Password2 string `json:"app_passwd2"`
		Active    string `json:"active"`
		Protocols string `json:"protocols"`
	}{
		Username:  username,
		AppName:   appname,
		Password:  password,
		Password2: password,
		Active:    "1",
		Protocols: protocols,
	}
	reqBytes, err := json.Marshal(reqStruct)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	req.Header.Add(ConstAPIKeyHeaderName, c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
