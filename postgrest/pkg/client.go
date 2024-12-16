package postgrest_go

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	session        http.Client
	Debug          bool
	defaultHeaders http.Header
	Transport      *PostgrestTransport
}

type ClientOption func(c *Client)

func NewClient(baseURL url.URL, opts ...ClientOption) *Client {
	transport := PostgrestTransport{
		baseURL: baseURL,
		Parent:  http.DefaultTransport,
	}

	c := Client{
		Transport:      &transport,
		defaultHeaders: http.Header{},
		session:        http.Client{Transport: &transport},
	}

	c.defaultHeaders.Set("Accept", "application/json")
	c.defaultHeaders.Set("Content-Type", "application/json")
	c.defaultHeaders.Set("Accept-Profile", "public")
	c.defaultHeaders.Set("Content-Profile", "public")

	for _, opt := range opts {
		opt(&c)
	}

	if c.Debug {
		fmt.Println("CAUTION! Please make sure to disable the debug option before deploying it to production.")
		c.Transport.debug = c.Debug
	}
	return &c
}

func (c *Client) From(table string) *RequestBuilder {
	return &RequestBuilder{
		client: c,
		path:   "/" + table,
		header: http.Header{},
		params: url.Values{},
	}
}

type RpcRequestBuilder struct {
	client     *Client
	path       string
	header     http.Header
	httpMethod string
	params     map[string]interface{}
}

func (c *Client) Rpc(f string, params map[string]interface{}) *RpcRequestBuilder {
	return &RpcRequestBuilder{
		client:     c,
		path:       c.Transport.baseURL.String() + "rpc/" + f,
		header:     http.Header{},
		httpMethod: http.MethodPost,
		params:     params,
	}
}

func (r *RpcRequestBuilder) Execute(result interface{}) error {
	return r.ExecuteWithContext(context.Background(), result)
}

func (r *RpcRequestBuilder) ExecuteWithContext(ctx context.Context, result interface{}) error {
	data, err := json.Marshal(r.params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, r.httpMethod, r.path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header = r.client.Headers()

	// inject/override custom headers
	for key, vals := range r.header {
		for _, val := range vals {
			req.Header.Set(key, val)
		}
	}

	req.URL.Path = req.URL.Path[1:]
	req.URL = r.client.Transport.baseURL.ResolveReference(req.URL)

	resp, err := r.client.session.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		reqError := RequestError{HTTPStatusCode: resp.StatusCode}

		if err = json.Unmarshal(body, &reqError); err != nil {
			return err
		}

		return &reqError
	}

	if resp.StatusCode != http.StatusNoContent && r != nil {
		if err = json.Unmarshal(body, result); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CloseIdleConnections() {
	c.session.CloseIdleConnections()
}

func (c *Client) Headers() http.Header {
	return c.defaultHeaders.Clone()
}

func (c *Client) AddHeader(key string, value string) {
	c.defaultHeaders.Set(key, value)
}

func WithTokenAuth(token string) ClientOption {
	return func(c *Client) {
		c.AddHeader("Authorization", "Bearer "+token)
	}
}

func WithBasicAuth(username, password string) ClientOption {
	return func(c *Client) {
		c.AddHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	}
}

func WithSchema(schema string) ClientOption {
	return func(c *Client) {
		c.AddHeader("Accept-Profile", schema)
		c.AddHeader("Content-Profile", schema)
	}
}
