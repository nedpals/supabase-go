package postgrest_go

import (
	"fmt"
	"net/http"
	"net/url"
)

type PostgrestTransport struct {
	baseURL url.URL
	debug   bool

	Parent http.RoundTripper
}

func (c *PostgrestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.debug {
		fmt.Println("--- incoming postgrest-go req ---")
		fmt.Printf("%s %s\n", req.Method, req.URL.String())
		for key, headerValues := range req.Header {
			for _, val := range headerValues {
				fmt.Printf("%s: %s\n", key, val)
			}
		}
		fmt.Println("---------------------------------")
	}

	return c.Parent.RoundTrip(req)
}
