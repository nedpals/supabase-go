package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	postgrest "github.com/nedpals/postgrest-go/pkg"
)

const (
	AuthEndpoint = "auth/v1"
	RestEndpoint = "rest/v1"
)

type Client struct {
	BaseURL string
	// apiKey can be a client API key or a service key
	apiKey     string
	HTTPClient *http.Client
	Auth       *Auth
	DB         *postgrest.Client
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (err *ErrorResponse) Error() string {
	return err.Message
}

// CreateClient creates a new Supabase client
func CreateClient(baseURL string, supabaseKey string, debug ...bool) *Client {
	parsedURL, err := url.Parse(fmt.Sprintf("%s/%s/", baseURL, RestEndpoint))
	if err != nil {
		panic(err)
	}
	client := &Client{
		BaseURL: baseURL,
		apiKey:  supabaseKey,
		Auth:    &Auth{},
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		DB: postgrest.NewClient(
			*parsedURL,
			postgrest.WithTokenAuth(supabaseKey),
			func(c *postgrest.Client) {
				// debug parameter is only for postgrest-go for now
				if len(debug) > 0 {
					c.Debug = debug[0]
				}
				c.AddHeader("apikey", supabaseKey)
			},
		),
	}
	client.Auth.client = client
	return client
}

func injectAuthorizationHeader(req *http.Request, value string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", value))
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	var errRes ErrorResponse

	req.Header.Set("apikey", c.apiKey)
	hasCustomError, err := c.sendCustomRequest(req, v, errRes)

	if err != nil {
		return err
	} else if hasCustomError {
		return &errRes
	}

	return nil
}

func (c *Client) sendCustomRequest(req *http.Request, successValue interface{}, errorValue interface{}) (bool, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return true, err
	}
	defer res.Body.Close()

	statusOK := res.StatusCode >= http.StatusOK && res.StatusCode < 300
	if !statusOK {
		if err = json.NewDecoder(res.Body).Decode(&errorValue); err == nil {
			return true, nil
		}

		return true, fmt.Errorf("unknown, status code: %d", res.StatusCode)
	}

	if res.StatusCode != http.StatusNoContent {
		if err = json.NewDecoder(res.Body).Decode(&successValue); err != nil {
			return true, err
		}
	}

	return false, nil
}
